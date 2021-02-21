package golive

import (
	"bytes"
	"fmt"
	"html/template"

	"golang.org/x/net/html"
)

type LiveState struct {
	html *html.Node
	text string
}

func (ls *LiveState) setText(text string) error {
	var err error
	ls.html, err = nodeFromString(text)
	ls.text = text
	return err
}

func (ls *LiveState) setHTML(node *html.Node) error {
	var err error
	ls.text, err = renderInnerHTML(node)
	ls.html = node
	return err
}

type LiveRenderer struct {
	state          *LiveState
	template       *template.Template
	templateString string
	formatters     []func(t string) string
}

func (lr *LiveRenderer) setTemplate(t *template.Template, ts string) {
	lr.template = t
	lr.templateString = ts
}

func (lr *LiveRenderer) renderToText(data interface{}) (string, error) {
	if lr.template == nil {
		return "", fmt.Errorf("template is not defined in LiveRenderer")
	}

	s := bytes.NewBufferString("")

	err := lr.template.Execute(s, data)

	if err != nil {
		err = fmt.Errorf("template execute: %w", err)
	}

	text := s.String()
	for _, f := range lr.formatters {
		text = f(text)
	}

	return text, err
}

func (lr *LiveRenderer) Render(data interface{}) (string, *html.Node, error) {

	textRender, err := lr.renderToText(data)
	if err != nil {
		return "", nil, err
	}

	err = lr.state.setText(textRender)
	return lr.state.text, lr.state.html, err
}

func (lr *LiveRenderer) LiveRender(data interface{}) (*diff, error) {

	actualRender := lr.state.html
	proposedRenderText, err := lr.renderToText(data)

	err = lr.state.setText(proposedRenderText)

	if err != nil {
		return nil, fmt.Errorf("state set text: %w", err)
	}

	// TODO: maybe the right way to call a diff is calling based on state
	diff := newDiff(actualRender)
	diff.propose(lr.state.html)

	return diff, nil
}

func (lr *LiveRenderer) useFormatter(f func(t string) string) {
	lr.formatters = append(lr.formatters, f)
}
