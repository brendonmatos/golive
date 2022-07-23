package renderer

import (
	"fmt"

	"github.com/brendonmatos/golive/live/component"
	"golang.org/x/net/html"
)

type StaticRenderer struct {
	template string
	handler  func(state *component.State) []interface{}
}

func NewStaticRenderer(t string, h func(state *component.State) []interface{}) *StaticRenderer {
	return &StaticRenderer{
		template: t,
		handler:  h,
	}
}

func (s *StaticRenderer) Prepare(state *State) error {
	s.template = signHtmlTemplate(s.template, state.Identifier)

	return nil
}

func (s *StaticRenderer) Render(state *component.State) (*string, *html.Node, error) {
	result := fmt.Sprintf(s.template, s.handler(state)...)

	return &result, nil, nil
}
