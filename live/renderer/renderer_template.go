package renderer

import (
	"bytes"
	"fmt"
	"github.com/brendonmatos/golive/live/state"
	"golang.org/x/net/html"
	"html/template"
	"regexp"
)

const GoLiveUidAttrKey = "gl-uid"

type TemplateRenderer struct {
	template       *template.Template
	templateString string
}

func (tr *TemplateRenderer) Prepare(state *State) {
	tr.templateString = signHtmlString(tr.templateString, state.Identifier)
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+[ ]?)`)

func (tr *TemplateRenderer) renderToText(data interface{}) (string, error) {
	if tr.template == nil {
		return "", fmt.Errorf("template is not defined in Renderer")
	}

	s := bytes.NewBufferString("")

	err := tr.template.Execute(s, data)

	if err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	text := s.String()

	return text, nil
}

func (tr *TemplateRenderer) Render(s *state.State) (string, *html.Node, error) {

	textRender, err := tr.renderToText(s.Value)
	if err != nil {
		return "", nil, err
	}

	return textRender, nil, nil
}
