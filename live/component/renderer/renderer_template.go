package renderer

import (
	"bytes"
	"fmt"
	state2 "github.com/brendonmatos/golive/live/component/state"
	"golang.org/x/net/html"
	"html/template"
)

const GoLiveUidAttrKey = "gl-uid"

type TemplateRenderer struct {
	template       *template.Template
	templateString string
	renderChild    RenderChild
}

func NewTemplateRenderer(templateStr string) *TemplateRenderer {
	return &TemplateRenderer{
		template:       nil,
		templateString: templateStr,
	}
}

func (tr *TemplateRenderer) SetRenderChild(fn RenderChild) (error, bool) {
	tr.renderChild = fn
	return nil, true
}

func (tr *TemplateRenderer) Prepare(state *State) error {
	tr.templateString = signHtmlTemplate(tr.templateString, state.Identifier)
	tpl := template.New(state.Identifier)

	tpl.Funcs(template.FuncMap{
		"render": func(st string) (*template.HTML, error) {
			renderer, err := tr.renderChild(st)

			if err != nil {
				return nil, err
			}

			t := template.HTML(renderer)
			return &t, nil
		},
	})

	parsed, err := tpl.Parse(tr.templateString)

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

func (tr *TemplateRenderer) Render(s *state2.State) (*string, *html.Node, error) {

	textRender, err := tr.renderToText(s.Value)
	if err != nil {
		return nil, nil, err
	}

	return &textRender, nil, nil
}