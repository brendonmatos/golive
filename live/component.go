package live

import (
	"errors"
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/renderer"
	"github.com/brendonmatos/golive/live/state"
	"github.com/brendonmatos/golive/live/util"
)

var (
	ErrComponentNotPrepared = errors.New("component need to be prepared")
	ErrComponentWithoutLog  = errors.New("component without log defined")
	ErrComponentNil         = errors.New("component nil")
)

type Component struct {
	Name     string
	Log      golive.Log
	Context  *Context
	State    *state.State
	Renderer *renderer.Renderer
}

func DefineComponent(name string) *Component {
	s := state.NewState()

	uid := util.CreateUniqueName(name)

	r := renderer.NewRenderer(renderer.NewTemplateRenderer(""))

	return &Component{
		Name:     uid,
		State:    s,
		Renderer: r,
		Context:  NewContext(),
	}
}

func (c *Component) CallHook(name string) {
	c.Context.CallHook(name)
}

func (c *Component) SetState(i interface{}) {
	c.State.Set(i)
}

func (c *Component) Create() error {
	var err error

	if c.Log == nil {
		return ErrComponentWithoutLog
	}

	c.CallHook("BeforeCreate")

	err = c.Renderer.Prepare(c.Name)

	if err != nil {
		return fmt.Errorf("renderer prepare: %w", err)
	}

	c.CallHook("Created")

	return err
}

func OnCreated(c *Component, h Hook) {
	fmt.Println("registering creation")
	c.Context.InjectHook("Created", h)
}

func OnUpdate(c *Component, h Hook) {
	c.Context.InjectHook("Update", h)
}

func (c *Component) UseRender(newRenderer renderer.RendererInterface) error {
	c.Renderer = renderer.NewRenderer(newRenderer)
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
	c.CallHook("Update")
}

// Unmount ...
func (c *Component) Unmount() error {
	c.Log(golive.LogTrace, "WillUnmount", golive.LogEx{"name": c.Name})
	c.CallHook("BeforeUnmount")
	c.State.Kill()
	c.CallHook("Unmounted")
	return nil
}
