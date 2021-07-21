package component

import (
	"errors"
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/component/renderer"
	"github.com/brendonmatos/golive/live/component/state"
	"github.com/brendonmatos/golive/live/util"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"reflect"
)

const GoLiveInput = "gl-input"

var (
	ErrComponentNotPrepared = errors.New("component need to be prepared")
	ErrComponentWithoutLog  = errors.New("component without log defined")
	ErrComponentNil         = errors.New("component nil")
)

const (
	BeforeMount   = "before_mount"
	Mounted       = "mounted"
	Update        = "update"
	Updated       = "updated"
	BeforeUnmount = "before_unmount"
	Unmounted     = "unmounted"
)

type Component struct {
	Name     string
	Log      golive.Log
	Context  *Context
	State    *state.State
	Renderer *renderer.Renderer

	componentsRegister map[string]interface{}
	children           []*Component
}

func NewLiveComponent(name string, state interface{}) *Component {
	c := DefineComponent(name)
	c.SetState(state)

	template, _ := c.State.InvokeMethodInPath("TemplateHandler", []reflect.Value{reflect.ValueOf(c)})
	c.UseRender(renderer.NewTemplateRenderer(template[0].String()))

	return c
}

func DefineComponent(name string) *Component {
	s := state.NewState()
	uid := util.CreateUniqueName(name)
	r := renderer.NewRenderer(renderer.NewTemplateRenderer(""))

	c := &Component{
		Name:     uid,
		State:    s,
		Renderer: r,
		Context:  nil,

		componentsRegister: map[string]interface{}{},
		children:           []*Component{},
	}

	c.SetContext(NewContext())

	r.UseFormatter(func(t *html.Node) {
		err := c.SignRender(t)
		if err != nil {
			panic(err)
		}
	})

	return c
}

func (c *Component) SignRender(node *html.Node) error {

	// Post treatment
	for _, node := range dom.GetAllChildrenRecursive(node, c.Name) {

		if goLiveInputAttr := dom.GetAttribute(node, GoLiveInput); goLiveInputAttr != nil {
			dom.AddNodeAttribute(node, ":value", goLiveInputAttr.Val)
		}

		if valueAttr := dom.GetAttribute(node, ":value"); valueAttr != nil {
			dom.RemoveNodeAttribute(node, ":value")

			f, err := c.State.GetFieldFromPath(valueAttr.Val)

			if err != nil {
				return err
			}

			if inputTypeAttr := dom.GetAttribute(node, "type"); inputTypeAttr != nil {
				switch inputTypeAttr.Val {
				case "checkbox":
					if f.Bool() {
						dom.AddNodeAttribute(node, "checked", "true")
					} else {
						dom.RemoveNodeAttribute(node, "checked")
					}
					break
				}
			} else if node.DataAtom == atom.Textarea {
				n, err := dom.NodeFromString(fmt.Sprintf("%v", f))

				if n == nil || n.FirstChild == nil {
					continue
				}

				if err != nil {
					return err
				}

				child := n.FirstChild

				n.RemoveChild(child)
				node.AppendChild(child)
			} else {
				dom.AddNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}

		if disabledAttr := dom.GetAttribute(node, ":disabled"); disabledAttr != nil {
			dom.RemoveNodeAttribute(node, ":disabled")
			if disabledAttr.Val == "true" {
				dom.AddNodeAttribute(node, "disabled", "")
			} else {
				dom.RemoveNodeAttribute(node, "disabled")
			}
		}
	}
	return nil
}

func (c *Component) CallHook(name string) error {
	return c.Context.CallHook(name)
}

func (c *Component) SetState(i interface{}) {
	c.State.Set(i)
}

func (c *Component) SetContext(ctx *Context) {
	ctx.Component = c
	c.Context = ctx
}

func (c *Component) Mount() error {
	var err error

	if c.Log == nil {
		return ErrComponentWithoutLog
	}

	err = c.CallHook(BeforeMount)
	if err != nil {
		return fmt.Errorf("before mount hook: %w", err)
	}

	err = c.Renderer.SetRenderChild(func(cn string) (string, error) {
		return c.RenderChild(cn, []reflect.Value{})
	})
	if err != nil {
		return fmt.Errorf("set render child: %w", err)
	}

	if err = c.Renderer.Prepare(c.Name); err != nil {
		return fmt.Errorf("renderer prepare: %w", err)
	}

	err = c.CallHook(Mounted)
	if err != nil {
		return fmt.Errorf("mounted: %w", err)
	}

	return err
}

func OnMounted(c *Component, h Hook) {
	c.Context.InjectHook(Mounted, h)
}

func OnUpdate(c *Component, h Hook) {
	c.Context.InjectHook(Update, h)
}

func (c *Component) UseRender(newRenderer renderer.RendererInterface) error {
	c.Renderer.Renderer = newRenderer
	return nil
}

// RenderStatic ...
func (c *Component) RenderStatic() (string, error) {
	c.Log(golive.LogDebug, "RenderStatic", golive.LogEx{"name": c.Name})

	text, _, err := c.Renderer.RenderState(c.State)
	return text, err
}

// LiveRender render a new version of the Component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (c *Component) LiveRender() (*differ.Diff, error) {
	return c.Renderer.RenderStateDiff(c.State)
}

func (c *Component) Update() {
	c.CallHook(Update)
}

// Unmount ...
func (c *Component) Unmount() error {
	c.Log(golive.LogTrace, "WillUnmount", golive.LogEx{"name": c.Name})
	c.CallHook(BeforeUnmount)
	c.State.Kill()
	err := c.Context.Close()
	if err != nil {
		return fmt.Errorf("close context: %w", err)
	}
	c.CallHook(Unmounted)
	return nil
}

func (c *Component) UseComponent(s string, cd interface{}) {
	c.componentsRegister[s] = cd
}

// SetupChild needs to be called just once per prop change
func (c *Component) SetupChild(s string, props []reflect.Value) (*Component, error) {
	cd, found := c.componentsRegister[s]

	if !found {
		return nil, errors.New("component not found")
	}
	v := reflect.ValueOf(cd)
	r := v.Call(props)
	cp := r[0].Interface().(*Component)

	ctx := c.Context.Child()
	cp.SetContext(ctx)
	cp.Log = c.Log
	c.children = append(c.children, cp)

	return cp, nil
}

func (c *Component) FindComponent(cid string) *Component {
	for _, child := range c.children {
		if child.Name == cid {
			return child
		}
	}

	for _, child := range c.children {

		found := child.FindComponent(cid)

		if found != nil {
			return found
		}
	}

	return nil
}

func (c *Component) RenderChild(s string, props []reflect.Value) (string, error) {

	cp, err := c.SetupChild(s, props)
	if err != nil {
		return "", fmt.Errorf("setup child: %w", err)
	}

	c.Log(golive.LogDebug, "SetupChild", golive.LogEx{"name": cp.Name})

	err = cp.Mount()
	if err != nil {
		return "", fmt.Errorf("mount child: %w", err)
	}

	c.Log(golive.LogDebug, "ChildMounted", golive.LogEx{"name": cp.Name})

	return cp.RenderStatic()
}
