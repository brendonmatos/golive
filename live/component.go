package live

import (
	"errors"
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/renderer"
	"github.com/brendonmatos/golive/live/state"
	"github.com/brendonmatos/golive/live/util"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"html/template"
	"reflect"
)

var (
	ErrComponentNotPrepared    = errors.New("component need to be prepared")
	ErrComponentWithoutLog     = errors.New("component without log defined")
	ErrComponentNil            = errors.New("component nil")
	ErrComponentNotFound       = errors.New("component not found")
	ErrComponentNotFoundToNode = errors.New("component not found to specified node")
)

type LifeTime interface {
	Create(component *Component)
	TemplateHandler(component *Component) string
	Mounted(component *Component)
	BeforeMount(component *Component)
	BeforeUnmount(component *Component)
}

type Child interface{}

type Component struct {
	Name string

	IsMounted bool
	IsCreated bool
	Exited    bool

	Log  golive.Log
	life *LifeCycle

	State    *state.State
	Renderer *renderer.Renderer

	children []*Component
}

func DefineComponent(name string) *Component {
	s := state.NewState()

	uid := util.CreateUniqueName(name)
	r := renderer.NewRenderer(uid, renderer.NewStaticRenderer("", func(state *state.State) []interface{} {
		return []interface{}{}
	}))
	return &Component{
		Name:     uid,
		State:    s,
		Renderer: r,
	}
}

// NewLiveComponent ...
func NewLiveComponent(name string, component interface{}) *Component {

	uid := util.CreateUniqueName(name)

	r := renderer.NewRenderer(uid, &renderer.TemplateRenderer{})

	s := state.NewState()

	s.Set(component)

	return &Component{
		Name:     uid,
		State:    s,
		Renderer: r,
	}
}

func (c *Component) SetState(i interface{}) {
	c.State.Set(i)
}

func (c *Component) Create(life *LifeCycle) error {
	var err error

	c.life = life

	if c.Log == nil {
		return ErrComponentWithoutLog
	}

	c.notifyStage(WillCreate)

	c.Renderer.UseFormatter(func(t *html.Node) {
		_ = c.treatRender(t)
	})

	c.Renderer.Prepare()

	// Calling Component creation
	c.MaybeInvokeInState("Create")

	err = c.createChildren()

	if err != nil {
		return err
	}

	c.IsCreated = true

	c.notifyStage(Created)

	return err
}

func (c *Component) MaybeInvokeInState(name string) {
	m := reflect.ValueOf(c.State.Value).MethodByName(name)

	if m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(c.State)})
	}
}

func (c *Component) FindComponentByID(id string) *Component {
	if c.Name == id {
		return c
	}

	for _, child := range c.children {
		if child.Name == id {
			return child
		}
	}

	for _, child := range c.children {
		found := child.FindComponentByID(id)

		if found != nil {
			return found
		}
	}

	return nil
}

// Mount 2. the Component loading html
func (c *Component) Mount() error {

	if !c.IsCreated {
		return ErrComponentNotPrepared
	}

	c.notifyStage(WillMount)

	c.MaybeInvokeInState("BeforeMount")

	err := c.MountChildren()

	if err != nil {
		return fmt.Errorf("mount children: %w", err)
	}

	c.MaybeInvokeInState("Mounted")

	c.IsMounted = true

	c.notifyStage(Mounted)

	return nil
}

func (c *Component) MountChildren() error {
	c.notifyStage(WillMountChildren)
	for _, child := range c.children {
		if !child.IsMounted {
			err := child.Mount()

			if err != nil {
				return fmt.Errorf("child mount error: %w", err)
			}
		}
	}
	c.notifyStage(ChildrenMounted)
	return nil
}

// Render ...
func (c *Component) Render() (string, error) {
	c.Log(golive.LogDebug, "Render", golive.LogEx{"name": c.Name})

	if c.State.Value == nil {
		return "", ErrComponentNil
	}

	text, _, err := c.Renderer.RenderState(c.State)
	return text, err
}

func (c *Component) RenderChild(fn reflect.Value, _ ...reflect.Value) (template.HTML, error) {
	var err error
	var childMaker func() *Component
	var child *Component

	toParse := fn.Interface()

	child, ok := toParse.(*Component)
	if ok {
		goto Render
	}

	childMaker, ok = toParse.(func() *Component)
	if ok {
		child = childMaker()

		if err := c.createChild(child); err != nil {
			return "", fmt.Errorf("create child: %w", err)
		}

		err := child.Mount()
		if err != nil {
			return "", fmt.Errorf("mount child: %w", err)
		}
	}

Render:
	c.Log(golive.LogDebug, "getting child result", golive.LogEx{"child": child})

	if child == nil {
		return "", nil
	}

	render, err := child.Render()
	if err != nil {
		c.Log(golive.LogError, "render child: render", golive.LogEx{"error": err})
	}

	return template.HTML(render), nil
}

// LiveRender render a new version of the Component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (c *Component) LiveRender() (*differ.Diff, error) {
	return c.Renderer.RenderStateDiff(c.State)
}

func (c *Component) Update() {
	c.notifyStage(Updated)
}

func (c *Component) UpdateWithSource(source *EventSource) {
	c.notifyStageWithSource(Updated, source)
}

// Kill ...
func (c *Component) Kill() error {

	c.KillChildren()

	c.Log(golive.LogTrace, "WillUnmount", golive.LogEx{"name": c.Name})

	c.MaybeInvokeInState("BeforeUnmount")

	c.notifyStage(WillUnmount)

	c.Exited = true
	c.State.Kill()

	c.notifyStage(Unmounted)

	c.life = nil

	return nil
}

func (c *Component) KillChildren() {
	for _, child := range c.children {
		if err := child.Kill(); err != nil {
			c.Log(golive.LogError, "kill child", golive.LogEx{"name": child.Name})
		}
	}
}

func (c *Component) createChild(child *Component) error {

	child.Log = c.Log
	err := child.Create(c.life)
	if err != nil {
		return err
	}
	c.children = append(c.children, child)
	return nil
}

func (c *Component) createChildren() error {
	var err error
	for _, child := range c.getChildrenComponents() {

		if child == nil {
			continue
		}

		err = c.createChild(child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Component) getChildrenComponents() []*Component {
	components := make([]*Component, 0)
	v := reflect.ValueOf(c.State).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		lc, ok := v.Field(i).Interface().(*Component)
		if !ok {
			continue
		}

		components = append(components, lc)
	}
	return components
}

func (c *Component) notifyStage(ltu LifeTimeStage) {
	c.notifyStageWithSource(ltu, nil)
}

func (c *Component) notifyStageWithSource(ltu LifeTimeStage, source *EventSource) {
	if c.life == nil {
		c.Log(golive.LogWarn, "Component life updates channel is nil", nil)
		return
	}

	*c.life <- LifeTimeMessage{
		Stage:     ltu,
		Component: c,
		Source:    source,
	}
}

func (c *Component) treatRender(dom *html.Node) error {

	// Post treatment
	for _, node := range differ.GetAllChildrenRecursive(dom) {

		if goLiveInputAttr := differ.GetAttribute(node, "gl-input"); goLiveInputAttr != nil {
			differ.AddNodeAttribute(node, ":value", goLiveInputAttr.Val)
		}

		if valueAttr := differ.GetAttribute(node, ":value"); valueAttr != nil {
			differ.RemoveNodeAttribute(node, ":value")

			cid, err := componentIDFromNode(node)

			if err != nil {
				return err
			}

			foundComponent := c.FindComponentByID(cid)

			if foundComponent == nil {
				return ErrComponentNotFound
			}

			f, err := foundComponent.State.GetFieldFromPath(valueAttr.Val)

			if err != nil {
				return nil
			}

			if inputTypeAttr := differ.GetAttribute(node, "type"); inputTypeAttr != nil {
				switch inputTypeAttr.Val {
				case "checkbox":
					if f.Bool() {
						differ.AddNodeAttribute(node, "checked", "checked")
					} else {
						differ.RemoveNodeAttribute(node, "checked")
					}
					break
				}
			} else if node.DataAtom == atom.Textarea {
				n, err := differ.NodeFromString(fmt.Sprintf("%v", f))

				if n == nil || n.FirstChild == nil {
					continue
				}

				if err != nil {
					continue
				}

				child := n.FirstChild

				n.RemoveChild(child)

				node.AppendChild(child)
			} else {
				differ.AddNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}

		if disabledAttr := differ.GetAttribute(node, ":disabled"); disabledAttr != nil {
			differ.RemoveNodeAttribute(node, ":disabled")
			if disabledAttr.Val == "true" {
				differ.AddNodeAttribute(node, "disabled", "")
			} else {
				differ.RemoveNodeAttribute(node, "disabled")
			}
		}
	}
	return nil
}

func componentIDFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := differ.GetAttribute(parent, differ.ComponentIdAttrKey); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", ErrComponentNotFoundToNode
}
