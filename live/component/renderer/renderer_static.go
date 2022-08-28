package renderer

import (
	"github.com/brendonmatos/golive/live/component"
	"golang.org/x/net/html"
)

type StaticRenderer struct {
	template string
}

func NewStaticRenderer(t string) *StaticRenderer {
	return &StaticRenderer{
		template: t,
	}
}

func (s *StaticRenderer) SetContent(t string) (string, error) {
	s.template = t
	return s.template, nil
}

func (s *StaticRenderer) Prepare(state *RenderState) error {
	s.SetContent(signHtmlTemplate(s.template, state.Identifier))

	return nil
}

func (s *StaticRenderer) Render(state *component.State) (*string, *html.Node, error) {

	return &s.template, nil, nil
}
