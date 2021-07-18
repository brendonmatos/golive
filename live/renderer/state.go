package renderer

import (
	"fmt"
	"github.com/brendonmatos/golive/differ"
	"golang.org/x/net/html"
)

type State struct {
	Identifier string
	html       *html.Node
	text       string
}

func NewRenderState(id string) *State {
	return &State{
		Identifier: id,
		html:       nil,
		text:       "",
	}
}

func (ls *State) SetText(text string) error {
	var err error
	ls.html, err = differ.NodeFromString(text)
	ls.text = text
	if err != nil {
		return fmt.Errorf("node from string: %w", err)
	}

	return nil
}

func (ls *State) SetHTML(node *html.Node) error {
	var err error
	ls.text, err = differ.RenderInnerHTML(node)
	ls.html = node
	if err != nil {
		return fmt.Errorf("render inner html: %w", err)
	}

	return nil
}
