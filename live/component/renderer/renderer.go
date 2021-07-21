package renderer

import (
	"errors"
	"fmt"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/dom"
	state2 "github.com/brendonmatos/golive/live/component/state"
	"golang.org/x/net/html"
)

var (
	ErrComponentNotFound       = errors.New("component not found")
	ErrComponentNotFoundToNode = errors.New("component not found to specified node")
)

type RenderChild func(string) (string, error)

type RendererInterface interface {
	SetRenderChild(child RenderChild) (error, bool)
	Prepare(state *State) error
	Render(state *state2.State) (*string, *html.Node, error)
}

type Renderer struct {
	State      *State
	Renderer   RendererInterface
	Formatters []func(t *html.Node)
}

func NewRenderer(r RendererInterface) *Renderer {
	return &Renderer{
		State:    nil,
		Renderer: r,
	}
}

func (r *Renderer) Prepare(id string) error {
	r.State = NewRenderState(id)
	return r.Renderer.Prepare(r.State)
}

func (r *Renderer) SetRenderChild(rc RenderChild) error {

	r.Renderer.SetRenderChild(rc)

	return nil
}

func (r *Renderer) RenderState(state *state2.State) (string, *html.Node, error) {

	renderString, renderHtml, err := r.Renderer.Render(state)

	if err != nil {
		return "", nil, fmt.Errorf("renderer render: %w", err)
	}

	if renderString != nil {
		err = r.State.SetText(*renderString)
		if err != nil {
			return "", nil, fmt.Errorf("set text: %w", err)
		}

		// Do state html job
		for _, f := range r.Formatters {
			f(r.State.html)
		}

		// TODO: optimize it. string to html to string
		err := r.State.SetHTML(r.State.html)
		if err != nil {
			return "", nil, err
		}

	}

	if renderHtml != nil {
		err := r.State.SetHTML(renderHtml)
		if err != nil {
			return "", nil, err
		}
	}

	return r.State.text, r.State.html, err
}

func (r *Renderer) RenderStateDiff(state *state2.State) (*differ.Diff, error) {

	actualRender := r.State.html

	_, newRenderHtml, err := r.RenderState(state)

	if err != nil {
		return nil, fmt.Errorf("render state: %w", err)
	}

	d := differ.NewDiff(actualRender)
	d.Propose(newRenderHtml)

	return d, nil
}

func (r *Renderer) UseFormatter(f func(t *html.Node)) {
	r.Formatters = append(r.Formatters, f)
}

func ComponentIDFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := dom.GetAttribute(parent, dom.ComponentIdAttrKey); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", ErrComponentNotFoundToNode
}
