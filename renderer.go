package golive

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"strconv"
)

type LiveState struct {
	html   *html.Node
	text	string
}

func (ls *LiveState) setText(text string) error {
	var err error
	ls.html, err = CreateDOMFromString(text)
	ls.text = text
	return err
}

func (ls *LiveState) setHTML(node *html.Node) error {
	var err error
	ls.text, err = RenderNodeToString(node)
	ls.html = node
	return err
}

type LiveRenderer struct {
	state LiveState
	template   *template.Template
}

func (lr *LiveRenderer) setTemplate(template *template.Template) {
	lr.template = template
}

func (lr *LiveRenderer) Render(data ...interface{}) (string, *html.Node, error) {
	if lr.template == nil {
		return "", nil, fmt.Errorf("template is not defined in LiveRenderer")
	}

	s := bytes.NewBufferString("")

	err := lr.template.Execute(s, data)



	if err != nil {
		return "", nil, err
	}

	err = lr.state.setText(s.String())

	return lr.state.text, lr.state.html, err
}

func (lr *LiveRenderer) LiveRender(data ...interface{}) (*PatchBrowser, error) {
	actualRenderTet :=

	om := NewPatchBrowser(l.Name)
	om.Name = EventLiveDom

	if lr.state.text == newRender {
		l.log(LogDebug, "render is identical with last", nil)
		return om, nil
	}


	changeInstructions, err := GetDiffFromRawHTML(l.rendered, newRender)

	if err != nil {
		l.log(LogPanic, "there is a error in diff", logEx{"error": err})
	}

	for _, instruction := range changeInstructions {

		selector, err := SelectorFromNode(instruction.Element)

		if err != nil {
			s, _ := RenderNodeToString(instruction.Element)
			l.log(LogPanic, "there is a error in selector", logEx{"error": err, "element": s})
		}

		om.AddInstruction(PatchInstruction{
			Name:     EventLiveDom,
			Type:     strconv.Itoa(int(instruction.Type)),
			Attr:     instruction.Attr,
			Content:  instruction.Content,
			Selector: selector,
		})
	}

	return om, nil
}

func sign(dom *html.Node, l *LiveComponent) {

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

		if liveInputParam, ok := attrs["go-live-input"]; ok {

			f := l.GetFieldFromPath(liveInputParam)

			if inputType, ok := attrs["type"]; ok && inputType == "checkbox" {
				if f.Bool() {
					addNodeAttribute(node, "checked", "checked")
				} else {
					removeNodeAttribute(node, "checked")
				}
			} else {
				addNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}
	}

}