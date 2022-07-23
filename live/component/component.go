package component

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/component/renderer"
	"github.com/brendonmatos/golive/live/util"
	html "github.com/levigross/exp-html"
	"github.com/levigross/exp-html/atom"
)

const GoLiveInput = "gl-input"

var (
	ErrComponentNotMounted    = errors.New("component need to be mounted")
	ErrComponentWithoutLog    = errors.New("component without log defined")
	ErrComponentNil           = errors.New("component nil")
	ErrChildComponentNotFound = errors.New("child component not found")
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
	State    *State
	Renderer *renderer.RenderController

	componentsRegister map[string]*Component
	children           []*Component

	Mounted bool
}

func UseState(ctx *Context, state interface{}) {
	ctx.Provided["state"] = state
	ctx.CallHook("set_state")
}

func DefineComponent(name string, setup func(ctx *Context) renderer.Renderer) *Component {

	uid := util.CreateUniqueName(name)
	ctx := NewContext()

	setUpRenderer := setup(ctx)

	rendererController := renderer.NewRenderer(setUpRenderer)

	c := &Component{
		Name:               uid,
		State:              nil,
		Renderer:           rendererController,
		Context:            ctx,
		componentsRegister: map[string]*Component{},
		children:           []*Component{},
	}

	ctx.SetHook("set_state", func(ctx *Context) {
		c.State.Set(ctx.Provided["set_state"])
	})

	rendererController.UseFormatter(func(t *html.Node) {
		err := c.SignRender(t)
		if err != nil {
			panic(err)
		}
	})

	c.SetContext(ctx)

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

func (c *Component) CallHook(name string) {
	err := c.Context.CallHook(name)

	if err != nil {
		c.Log(golive.LogError, fmt.Sprintf("call hook error: %s", err), golive.LogEx{})
	}
}

func (c *Component) SetState(i interface{}) {
	c.State.Set(i)
}

func (c *Component) SetContext(ctx *Context) {
	ctx.Component = c

	if c.Context != nil {
		for s, hooks := range c.Context.Hooks {
			for _, hook := range hooks {
				ctx.SetHook(s, hook)
			}
		}
	}

	c.Context = ctx
}

func (c *Component) Mount() error {
	var err error
	if c.Log == nil {
		return ErrComponentWithoutLog
	}
	c.CallHook(BeforeMount)
	err = c.Renderer.SetRenderChild(func(cn string) (string, error) {
		return c.RenderChild(cn, []reflect.Value{})
	})
	if err != nil {
		return fmt.Errorf("set render child: %w", err)
	}
	if err = c.Renderer.Prepare(c.Name); err != nil {
		return fmt.Errorf("renderer prepare: %w", err)
	}
	c.CallHook(Mounted)
	c.Mounted = true
	return err
}

func OnMounted(c *Component, h Hook) {
	c.Context.SetHook(Mounted, h)
}

func OnBeforeMount(c *Component, h Hook) {
	c.Context.SetHook(BeforeMount, h)
}

func OnUpdate(c *Component, h Hook) {
	c.Context.SetHook(Update, h)
}

// RenderStatic ...
func (c *Component) RenderStatic() (string, error) {
	if !c.Mounted {
		return "", ErrComponentNotMounted
	}

	c.Log(golive.LogDebug, "RenderStatic", golive.LogEx{"name": c.Name})

	text, _, err := c.Renderer.RenderState(c.State)
	return text, err
}

// LiveRender render a new version of the Component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (c *Component) LiveRender() (*differ.Diff, error) {
	if !c.Mounted {
		return nil, ErrComponentNotMounted
	}

	return c.Renderer.RenderStateDiff(c.State)
}

func (c *Component) Update() {
	c.CallHook(Update)
}

// Unmount ...
func (c *Component) Unmount() error {
	if !c.Mounted {
		return ErrComponentNotMounted
	}
	c.Log(golive.LogTrace, "WillUnmount", golive.LogEx{"name": c.Name})
	c.CallHook(BeforeUnmount)
	c.State.Kill()
	err := c.Context.Close()
	if err != nil {
		return fmt.Errorf("context close: %w", err)
	}
	c.CallHook(Unmounted)
	return nil
}

func (c *Component) UseComponent(s string, cd *Component) {
	c.componentsRegister[s] = cd
}

// SetupChild needs to be called just once per prop change
func (c *Component) SetupChild(s string, props []reflect.Value) (*Component, error) {
	cd, found := c.componentsRegister[s]

	if !found {
		return nil, fmt.Errorf("component are not registered: %s", s)
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

func (c *Component) FindComponent(cid string) (*Component, error) {
	for _, child := range c.children {
		if child.Name == cid {
			return child, nil
		}
	}

	for _, child := range c.children {

		found, _ := child.FindComponent(cid)

		if found != nil {
			return found, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrChildComponentNotFound, cid)
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

func Provide(c *Component, symbol string, value interface{}) {
	r := c.Context.Root
	r.Provided[symbol] = value
}

func Inject(c *Component, symbol string) interface{} {
	r := c.Context.Root
	return r.Provided[symbol]
}
