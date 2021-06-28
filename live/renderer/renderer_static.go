package renderer

import (
	"fmt"
	"github.com/brendonmatos/golive/live/state"
	"golang.org/x/net/html"
)

type StaticRenderer struct {
	Template string
	Handler  func(state *state.State) []interface{}
}

func (s *StaticRenderer) Prepare(state *State) {
	s.Template = signHtmlString(s.Template, state.Identifier)
}

func NewStaticRenderer(t string, h func(state *state.State) []interface{}) *StaticRenderer {
	return &StaticRenderer{
		Template: t,
		Handler:  h,
	}
}

func (s *StaticRenderer) Render(state *state.State) (string, *html.Node, error) {
	result := fmt.Sprintf(s.Template, s.Handler(state)...)

	return result, nil, nil
}
