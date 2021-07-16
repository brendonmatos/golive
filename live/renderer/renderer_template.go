package renderer

import (
	"bytes"
	"fmt"
	"github.com/brendonmatos/golive/live/state"
	"golang.org/x/net/html"
	"html/template"
)

const GoLiveUidAttrKey = "gl-uid"

type TemplateRenderer struct {
	template       *template.Template
	templateString string
}

func NewTemplateRenderer(templateStr string) *TemplateRenderer {
	return &TemplateRenderer{
		template:       nil,
		templateString: templateStr,
	}
}

func (tr *TemplateRenderer) Prepare(state *State) error {
	tr.templateString = signHtmlTemplate(tr.templateString, state.Identifier)
	parsed, err := template.New(state.Identifier).Parse(tr.templateString)

	tr.template = parsed

	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	return nil
}

func (tr *TemplateRenderer) SetTemplate(template string) error {
	tr.templateString = template

	return nil
}

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

func (tr *TemplateRenderer) Render(s *state.State) (*string, *html.Node, error) {

	textRender, err := tr.renderToText(s.Value)
	if err != nil {
		return nil, nil, err
	}

	return &textRender, nil, nil
}
