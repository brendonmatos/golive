package renderer

import (
	"errors"

	"github.com/brendonmatos/golive/dom"

	"golang.org/x/net/html"
)

var (
	ErrComponentNotFound       = errors.New("component not found")
	ErrComponentNotFoundToNode = errors.New("component not found to specified node")
)

// type RenderChild func(string) (string, error)

// type Renderer interface {
// 	SetRenderChild(child RenderChild) (error, bool)
// 	Prepare(state *RenderState) error
// 	Render(state *RenderState) (*string, *html.Node, error)
// }

// type RenderController struct {
// 	State      *RenderState
// 	Formatters []func(t *html.Node)
// }

// func NewRendererController() *RenderController {
// 	return &RenderController{
// 		State: nil,
// 	}
// }

// func (r *RenderController) Prepare(id string) error {
// 	r.State = NewRenderState(id)
// 	return r.Renderer.Prepare(r.State)
// }

// func (r *RenderController) SetRenderChild(rc RenderChild) error {

// 	r.Renderer.SetRenderChild(rc)

// 	return nil
// }

// func (r *RenderController) RenderState(state *RenderState) (string, *html.Node, error) {

// 	renderString, renderHtml, err := r.Renderer.Render(state)

// 	if err != nil {
// 		return "", nil, fmt.Errorf("renderer render: %w", err)
// 	}

// 	if renderString != nil {
// 		err = r.State.SetText(*renderString)
// 		if err != nil {
// 			return "", nil, fmt.Errorf("set text: %w", err)
// 		}

// 		// Do state html job
// 		for _, f := range r.Formatters {
// 			f(r.State.html)
// 		}

// 		// TODO: optimize it. string to html to string
// 		err := r.State.SetHTML(r.State.html)
// 		if err != nil {
// 			return "", nil, err
// 		}

// 	}

// 	if renderHtml != nil {
// 		err := r.State.SetHTML(renderHtml)
// 		if err != nil {
// 			return "", nil, err
// 		}
// 	}

// 	return r.State.text, r.State.html, err
// }

// func (r *RenderController) RenderStateDiff(state *RenderState) (*differ.Diff, error) {

// 	actualRender := r.State.html

// 	_, newRenderHtml, err := r.RenderState(state)

// 	if err != nil {
// 		return nil, fmt.Errorf("render state: %w", err)
// 	}

// 	d := differ.NewDiff(actualRender)
// 	d.Propose(newRenderHtml)

// 	return d, nil
// }

// func (r *RenderController) UseFormatter(f func(t *html.Node)) {
// 	r.Formatters = append(r.Formatters, f)
// }

func ComponentIDFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := dom.GetAttribute(parent, dom.ComponentIdAttrKey); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", ErrComponentNotFoundToNode
}
