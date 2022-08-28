package renderer

const GoLiveUidAttrKey = "gl-uid"

// type TemplateRenderer struct {
// 	template       *template.Template
// 	templateString string
// 	renderChild    RenderChild
// 	funcs          []func(*RenderState) template.FuncMap
// }

// func NewTemplateRenderer(templateStr string) Renderer {
// 	return &TemplateRenderer{
// 		template:       nil,
// 		templateString: templateStr,
// 		funcs:          []func(*RenderState) template.FuncMap{},
// 	}
// }

// func (tr *TemplateRenderer) SetRenderChild(fn RenderChild) (error, bool) {
// 	tr.renderChild = fn
// 	return nil, true
// }

// func (tr *TemplateRenderer) Prepare(state *RenderState) error {
// 	tr.templateString = signHtmlTemplate(tr.templateString, state.Identifier)
// 	tpl := template.New(state.Identifier)

// 	tpl.Funcs(template.FuncMap{
// 		"render": func(st string) (*template.HTML, error) {
// 			rendered, err := tr.renderChild(st)

// 			if err != nil {
// 				return nil, err
// 			}

// 			t := template.HTML(rendered)
// 			return &t, nil
// 		},
// 	})

// 	if tr.funcs != nil && len(tr.funcs) > 0 {
// 		for _, funcs := range tr.funcs {
// 			tpl.Funcs(funcs(state))
// 		}
// 	}

// 	parsed, err := tpl.Parse(tr.templateString)

// 	tr.template = parsed

// 	if err != nil {
// 		return fmt.Errorf("prepare: %w", err)
// 	}

// 	return nil
// }

// func (tr *TemplateRenderer) SetTemplate(template string) error {
// 	tr.templateString = template

// 	return nil
// }

// func (tr *TemplateRenderer) SetFuncs(funcs ...func(*RenderState) template.FuncMap) error {
// 	tr.funcs = funcs

// 	return nil
// }

// func (tr *TemplateRenderer) renderToText(data interface{}) (string, error) {
// 	if tr.template == nil {
// 		return "", fmt.Errorf("template is not defined in Renderer")
// 	}

// 	s := bytes.NewBufferString("")

// 	err := tr.template.Execute(s, data)

// 	if err != nil {
// 		return "", fmt.Errorf("template execute: %w", err)
// 	}

// 	text := s.String()

// 	return text, nil
// }

// func (tr *TemplateRenderer) Render(s *component.State) (*string, *html.Node, error) {

// 	textRender, err := tr.renderToText(s.Value)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	return &textRender, nil, nil
// }
