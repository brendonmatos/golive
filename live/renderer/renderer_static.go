package renderer

import (
	"fmt"
	"github.com/brendonmatos/golive/live/state"
	"golang.org/x/net/html"
)

type StaticRenderer struct {
	template string
	handler  func(state *state.State) []interface{}
}

func NewStaticRenderer(t string, h func(state *state.State) []interface{}) *StaticRenderer {
	return &StaticRenderer{
		template: t,
		handler:  h,
	}
}

func (s *StaticRenderer) Prepare(state *State) error {
	s.template = signHtmlTemplate(s.template, state.Identifier)

	return nil
}

func (s *StaticRenderer) Render(state *state.State) (*string, *html.Node, error) {
	result := fmt.Sprintf(s.template, s.handler(state)...)

	return &result, nil, nil
}
