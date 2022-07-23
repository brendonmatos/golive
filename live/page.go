package live

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/wire"
)

var BasePage *template.Template

//go:embed page.html
var BasePageString string

func init() {
	var err error
	BasePage, err = template.New("BasePage").Parse(BasePageString)
	if err != nil {
		panic(err)
	}
}

type PageEnum struct {
	EventLiveInput          wire.EventKind
	EventLiveMethod         wire.EventKind
	EventLiveDom            wire.EventKind
	EventLiveConnectElement wire.EventKind
	EventLiveError          wire.EventKind
	EventLiveNavigate       wire.EventKind
	DiffSetAttr             differ.Type
	DiffRemoveAttr          differ.Type
	DiffReplace             differ.Type
	DiffRemove              differ.Type
	DiffSetInnerHTML        differ.Type
	DiffAppend              differ.Type
	DiffMove                differ.Type
}

type Page struct {
	content        PageContent
	EntryComponent *component.Component

	// Components is a list that handle all the componentsRegister from the page
	Components map[string]*component.Component
}

type PageContent struct {
	Lang          string
	Body          template.HTML
	Head          template.HTML
	Script        string
	Title         string
	Enum          PageEnum
	EnumLiveError map[string]string
}

func NewLivePage(c *component.Component) *Page {
	return &Page{
		EntryComponent: c,
		Components:     make(map[string]*component.Component),
	}
}

func (lp *Page) SetContent(c PageContent) {
	lp.content = c
}

func (lp *Page) Render() (string, error) {
	rendered, err := lp.EntryComponent.RenderStatic()

	if err != nil {
		return "", fmt.Errorf("entry component render: %w", err)
	}

	// Body content
	lp.content.Body = template.HTML(rendered)
	lp.content.Enum = PageEnum{
		EventLiveInput:          wire.FromBrowserLiveInput,
		EventLiveMethod:         wire.FromBrowserLiveMethod,
		EventLiveDom:            wire.FromServerLiveDom,
		EventLiveError:          wire.FromServerLiveError,
		EventLiveConnectElement: wire.FromServerLiveConnectElement,
		EventLiveNavigate:       wire.FromServerLiveNavigate,
		DiffSetAttr:             differ.SetAttr,
		DiffRemoveAttr:          differ.RemoveAttr,
		DiffReplace:             differ.Replace,
		DiffRemove:              differ.Remove,
		DiffSetInnerHTML:        differ.SetInnerHTML,
		DiffAppend:              differ.Append,
		DiffMove:                differ.Move,
	}
	lp.content.EnumLiveError = ErrorMap()

	writer := bytes.NewBuffer([]byte{})
	err = BasePage.Execute(writer, lp.content)
	return writer.String(), err
}
