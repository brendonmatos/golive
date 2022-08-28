package renderer

import (
	"fmt"

	"github.com/brendonmatos/golive/dom"
	"golang.org/x/net/html"
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
