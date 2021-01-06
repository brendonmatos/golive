package golive

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"strconv"
)

type LiveState struct {
	html *html.Node
	text string
}

func (ls *LiveState) setText(text string) error {
	var err error
	ls.html, err = CreateDOMFromString(text)
	ls.text = text
	return err
}

func (ls *LiveState) setHTML(node *html.Node) error {
	var err error
	ls.text, err = RenderChildren(node)
	ls.html = node
	return err
}

type LiveRenderer struct {
	state          *LiveState
	template       *template.Template
	templateString string
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
		return "", err
	}

	return s.String(), nil
}

func (lr *LiveRenderer) Render(data interface{}) (string, *html.Node, error) {

	textRender, err := lr.renderToText(data)

	if err != nil {
		return "", nil, err
	}

	err = lr.state.setText(textRender)

	return lr.state.text, lr.state.html, err
}

func (lr *LiveRenderer) LiveRender(data interface{}) (*Diff, error) {

	actualRender := lr.state.html
	actualRenderText := lr.state.text
	proposedRenderText, err := lr.renderToText(data)

	if err != nil {
		return nil, err
	}

	diff := NewDiff(actualRender)

	if actualRenderText == proposedRenderText {
		return diff, nil
	}

	_ = lr.state.setText(proposedRenderText)

	diff.Propose(lr.state.html)

	return diff, nil
}

func signPreRender(dom *html.Node, l *LiveComponent) {
	// Post treatment
	for index, node := range GetAllChildrenRecursive(dom) {
		addNodeAttribute(node, "go-live-uid", l.Name+"_"+strconv.FormatInt(int64(index), 16))
	}
}

func signPostRender(dom *html.Node, l *LiveComponent) {

	// Post treatment
	for index, node := range GetAllChildrenRecursive(dom) {

		attrs := AttrMapFromNode(node)

		addNodeAttribute(node, "go-live-uid", l.Name+"_"+strconv.FormatInt(int64(index), 16))

		if isElementDisabled, ok := attrs[":disabled"]; ok {
			if isElementDisabled == "true" {
				addNodeAttribute(node, "disabled", "disabled")
			} else {
				removeNodeAttribute(node, "disabled")
			}
		}

		if goLiveInputParam, ok := attrs["go-live-input"]; ok {
			f := l.GetFieldFromPath(goLiveInputParam)
			if inputType, ok := attrs["type"]; ok && inputType == "checkbox" {
				if f.Bool() {
					addNodeAttribute(node, "checked", "checked")
				} else {
					removeNodeAttribute(node, "checked")
				}
			} else {
				addNodeAttribute(node, "value", goLiveInputParam)
			}
		}
	}

}
