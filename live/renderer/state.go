package renderer

import (
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
	return err
}

func (ls *State) SetHTML(node *html.Node) error {
	var err error
	ls.text, err = differ.RenderInnerHTML(node)
	ls.html = node
	return err
}
