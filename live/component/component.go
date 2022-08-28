package component

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/util"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	Mount         = "mount"
	Mounted       = "mounted"
	Update        = "update"
	Updated       = "updated"
	BeforeUnmount = "before_unmount"
	Unmounted     = "unmounted"
	SetState      = "set_state"
	Render        = "render"
	Rendered      = "rendered"
)

type RenderState struct {
	Identifier string
	html       *html.Node
	text       string
}

func NewRenderState(id string) *RenderState {
	return &RenderState{
		Identifier: id,
		html:       nil,
		text:       "",
	}
}

func (ls *RenderState) SetText(text string) error {
	var err error
	ls.html, err = dom.NodeFromString(text)
	ls.text = text
	if err != nil {
		return fmt.Errorf("node from string: %w", err)
	}

	return nil
}

func (ls *RenderState) SetHTML(node *html.Node) error {
	var err error
	ls.text, err = dom.RenderInnerHTML(node)
	ls.html = node
	if err != nil {
		return fmt.Errorf("render inner html: %w", err)
	}

	return nil
}

func (ls *RenderState) GetHTML() *html.Node {
	return ls.html
}

func (ls *RenderState) GetText() string {
	return ls.text
}

type Component struct {
	Name        string
	Log         golive.Log
	Context     *Context
	State       *State
	RenderState *RenderState

	componentsRegister map[string]*Component
	children           []*Component
	Props              interface{}
	Mounted            bool
}

func UseState[T interface{}](ctx *Context, state T) T {
	if ctx.Frozen {
		state, ok := ctx.Provided["state"].(T)
		if !ok {
			panic("invalid state")
		}
		return state
	}
	ctx.Provided["state"] = state
	ctx.CallHook(SetState)
	return state
}

func UseUpdate(ctx *Context) func() {
	return func() {
		ctx.CallHook(Update)
	}
}

func DefineComponent[T interface{}](name string, setup func(ctx *Context, props *T) string) *Component {

	uid := util.CreateUniqueName(name)
	ctx := NewContext()

	c := &Component{
		Name:               uid,
		State:              NewState(),
		Context:            ctx,
		componentsRegister: map[string]*Component{},
		children:           []*Component{},
		Props:              nil,
	}

	ctx.SetHook(SetState, func(ctx *Context) {
		c.State.Set(ctx.Provided["state"])
	})

	ctx.SetHook(Render, func(ctx *Context) {
		var render string
		if ctx.Component.Props == nil {
			render = setup(ctx, nil)
		} else {
			props := ctx.Component.Props.(*T)
			render = setup(ctx, props)
		}

		ctx.Frozen = true
		render = signHtmlTemplate(render, c.Name)
		rs := NewRenderState(uid)
		rs.SetText(render)
		c.SignRender(rs.GetHTML())
		ctx.CallHook(Rendered)
		c.RenderState = rs
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

func OnMounted(c *Component, h Hook) {
	c.Context.SetHook(Mounted, h)
}

func UseOnMounted(ctx *Context, cb func()) {
	ctx.SetHook(Mounted, func(ctx *Context) {
		cb()
	})
}

func OnBeforeMount(c *Component, h Hook) {
	c.Context.SetHook(BeforeMount, h)
}

func OnUpdate(c *Component, h Hook) {
	c.Context.SetHook(Update, h)
}

func (c *Component) Render() error {
	if !c.Mounted {
		return ErrComponentNotMounted
	}
	c.CallHook(Render)
	c.Log(golive.LogDebug, "Render", golive.LogEx{"name": c.Name})
	return nil
}

func (c *Component) RenderWithProps(props interface{}) string {
	c.Props = props
	c.Render()
	return c.RenderState.GetText()
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
	c.CallHook(Unmounted)
	err := c.Context.Close()
	if err != nil {
		return fmt.Errorf("context close: %w", err)
	}
	return nil
}

func (c *Component) UseComponent(s string, cd *Component) {
	c.componentsRegister[s] = cd
}
func (c *Component) Mount() error {
	var err error
	if c.Log == nil {
		return ErrComponentWithoutLog
	}
	c.CallHook(BeforeMount)
	c.CallHook(Render)
	c.CallHook(Mounted)
	c.Mounted = true
	return err
}

func (c *Component) FindComponent(cid string) (*Component, error) {

	if c.Name == cid {
		return c, nil
	}

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

func Provide(c *Component, symbol string, value interface{}) {
	r := c.Context.Root
	r.Provided[symbol] = value
}

func Inject(c *Component, symbol string) interface{} {
	r := c.Context.Root
	return r.Provided[symbol]
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+[ ]?)`)

func replaceWithFunction(content string, r *regexp.Regexp, h func(string) string) string {
	matches := r.FindAllStringSubmatchIndex(content, -1)

	util.ReverseSlice(matches)

	for _, match := range matches {
		startIndex := match[0]
		endIndex := match[1]

		startSlice := content[:startIndex]
		endSlide := content[endIndex:]
		matchedSlice := content[startIndex:endIndex]

		content = startSlice + h(matchedSlice) + endSlide
	}

	return content
}

const GoLiveUidAttrKey = "gl-uid"

func signHtmlTemplate(template string, uid string) string {

	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` ` + dom.ComponentIdAttrKey + `="` + uid + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}

	// template = replaceWithFunction(template, rxTagName, func(s string) string {
	// 	lUid := uid + "_" + util.RandomSmall()
	// 	return s + ` ` + GoLiveUidAttrKey + `="` + lUid + `" `
	// })

	return template
}
