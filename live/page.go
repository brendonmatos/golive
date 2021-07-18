package live

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/brendonmatos/golive/differ"
	"html/template"
	"reflect"
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
	EventLiveInput          string
	EventLiveMethod         string
	EventLiveDom            string
	EventLiveConnectElement string
	EventLiveError          string
	DiffSetAttr             differ.Type
	DiffRemoveAttr          differ.Type
	DiffReplace             differ.Type
	DiffRemove              differ.Type
	DiffSetInnerHTML        differ.Type
	DiffAppend              differ.Type
	DiffMove                differ.Type
}

type LivePageEvent struct {
	Type      int
	Component *Component
	Source    *EventSource
}

type LiveEventsChannel chan LivePageEvent

type Page struct {
	content             PageContent
	Events              LiveEventsChannel
	ComponentsLifeCycle *LifeCycle

	EntryComponent *Component

	// Components is a list that handle all the components from the page
	Components map[string]*Component
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

func NewLivePage(c *Component) *Page {
	componentsUpdatesChannel := make(LifeCycle)
	pageEventsChannel := make(LiveEventsChannel)

	return &Page{
		EntryComponent:      c,
		Events:              pageEventsChannel,
		ComponentsLifeCycle: &componentsUpdatesChannel,
		Components:          make(map[string]*Component),
	}
}

func (lp *Page) SetContent(c PageContent) {
	lp.content = c
}

// Create main component from page in sequence of life cycle
func (lp *Page) Create() {

	ctx := lp.EntryComponent.Context

	ctx.InjectGlobalHook("Created", func() {
		lp.Emit(PageComponentMounted, lp.EntryComponent)
	})

	ctx.InjectGlobalHook("Update", func() {
		lp.Emit(PageComponentUpdated, lp.EntryComponent)
	})

	// pass mount live Component with lifecycle channel
	err := lp.EntryComponent.Create()

	if err != nil {
		panic(fmt.Errorf("mount: create entryComponent: %w", err))
	}

}

func (lp *Page) Render() (string, error) {
	rendered, err := lp.EntryComponent.RenderStatic()

	if err != nil {
		return "", fmt.Errorf("entry component render: %w", err)
	}

	// Body content
	lp.content.Body = template.HTML(rendered)
	lp.content.Enum = PageEnum{
		EventLiveInput:          EventLiveInput,
		EventLiveMethod:         EventLiveMethod,
		EventLiveDom:            EventLiveDom,
		EventLiveError:          EventLiveError,
		EventLiveConnectElement: EventLiveConnectElement,
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

func (lp *Page) Emit(lts int, c *Component) {
	lp.EmitWithSource(lts, c, nil)
}

func (lp *Page) EmitWithSource(lts int, c *Component, source *EventSource) {
	if c == nil {
		c = lp.EntryComponent
	}

	lp.Events <- LivePageEvent{
		Type:      lts,
		Component: lp.EntryComponent,
		Source:    source,
	}
}

func (lp *Page) HandleBrowserEvent(m BrowserEvent) error {

	c := lp.EntryComponent

	if c == nil {
		return fmt.Errorf("Component not found with id: %s", m.ComponentID)
	}

	var err error
	switch m.Name {
	case EventLiveInput:
		err = c.State.SetValueInPath(m.StateValue, m.StateKey)
	case EventLiveMethod:
		err = c.State.InvokeMethodInPath(m.MethodName, []reflect.Value{reflect.ValueOf(m.MethodData), reflect.ValueOf(m.DOMEvent)})
	case EventLiveDisconnect:
		err = c.Unmount()
	}

	c.Update()

	return err
}

const PageComponentUpdated = 1
const PageComponentMounted = 2
