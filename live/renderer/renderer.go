package renderer

import (
	"fmt"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/state"
	"golang.org/x/net/html"
)

type Type interface {
	Prepare(state *State)
	Render(state *state.State) (string, *html.Node, error)
}

type Renderer struct {
	State      *State
	Renderer   Type
	Formatters []func(t *html.Node)
}

func NewRenderer(id string, r Type) *Renderer {
	return &Renderer{
		State:    NewRenderState(id),
		Renderer: r,
	}
}

func (r *Renderer) Prepare() {
	r.Renderer.Prepare(r.State)
}

func (r *Renderer) RenderState(state *state.State) (string, *html.Node, error) {

	renderString, renderHtml, err := r.Renderer.Render(state)

	if err != nil {
		return "", nil, fmt.Errorf("renderer render: %w", err)
	}

	if renderHtml != nil {
		err := r.State.SetHTML(renderHtml)

		if err != nil {
			return "", nil, err
		}
	}

	err = r.State.SetText(renderString)

	if err != nil {
		return "", nil, fmt.Errorf("set text: %w", err)
	}

	// Do state html job
	for _, f := range r.Formatters {
		f(r.State.html)
	}

	return r.State.text, r.State.html, err
}

func (r *Renderer) RenderStateDiff(state *state.State) (*differ.Diff, error) {

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
