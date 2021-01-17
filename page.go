package golive

import (
	"bytes"
	"fmt"
	"html/template"
)

var BasePage *template.Template

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
	DiffSetAttr             DiffType
	DiffRemoveAttr          DiffType
	DiffReplace             DiffType
	DiffRemove              DiffType
	DiffSetInnerHTML        DiffType
	DiffAppend              DiffType
}

type LivePageEvent struct {
	Type      int
	Component *LiveComponent
	Source    *EventSource
}

type LiveEventsChannel chan LivePageEvent

type Page struct {
	content             PageContent
	Events              LiveEventsChannel
	ComponentsLifeCycle *ComponentLifeCycle

	entry *LiveComponent

	// Components is a list that handle all the components from the page
	Components map[string]*LiveComponent
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

func NewLivePage(c *LiveComponent) *Page {
	componentsUpdatesChannel := make(ComponentLifeCycle)
	pageEventsChannel := make(LiveEventsChannel)

	return &Page{
		entry:               c,
		Events:              pageEventsChannel,
		ComponentsLifeCycle: &componentsUpdatesChannel,
		Components:          make(map[string]*LiveComponent),
	}
}

func (lp *Page) SetContent(c PageContent) {
	lp.content = c
}

// Call the component in sequence of life cycle
func (lp *Page) Mount() {

	// Enable components lifecycle channel receiver
	lp.enableComponentLifeCycleReceiver()

	// pass mount live component with lifecycle channel
	err := lp.entry.Create(lp.ComponentsLifeCycle)

	if err != nil {
		panic(fmt.Errorf("mount: create entry: %w", err))
	}

	err = lp.entry.Mount()

	if err != nil {
		panic(err)
	}

}

func (lp *Page) Render() (string, error) {
	// Render entry component
	rendered, err := lp.entry.Render()

	if err != nil {
		return "", err
	}

	// Body content
	lp.content.Body = template.HTML(rendered)
	lp.content.Enum = PageEnum{
		EventLiveInput:          EventLiveInput,
		EventLiveMethod:         EventLiveMethod,
		EventLiveDom:            EventLiveDom,
		EventLiveError:          EventLiveError,
		EventLiveConnectElement: EventLiveConnectElement,
		DiffSetAttr:             SetAttr,
		DiffRemoveAttr:          RemoveAttr,
		DiffReplace:             Replace,
		DiffRemove:              Remove,
		DiffSetInnerHTML:        SetInnerHtml,
		DiffAppend:              Append,
	}
	lp.content.EnumLiveError = LiveErrorMap()

	writer := bytes.NewBuffer([]byte{})
	err = BasePage.Execute(writer, lp.content)
	return writer.String(), err
}

func (lp *Page) Emit(lts int, c *LiveComponent) {
	lp.EmitWithSource(lts, c, nil)
}

func (lp *Page) EmitWithSource(lts int, c *LiveComponent, source *EventSource) {
	if c == nil {
		c = lp.entry
	}

	lp.Events <- LivePageEvent{
		Type:      lts,
		Component: c,
		Source:    source,
	}
}

func (lp *Page) HandleBrowserEvent(m BrowserEvent) error {

	c := lp.entry.findComponentByID(m.ComponentID)

	if c == nil {
		return fmt.Errorf("component not found with id: %s", m.ComponentID)
	}

	var source *EventSource

	var err error
	switch m.Name {
	case EventLiveInput:
		err = c.SetValueInPath(m.StateValue, m.StateKey)
		source = &EventSource{Type: EventSourceInput, Value: m.StateKey}
	case EventLiveMethod:
		err = c.InvokeMethodInPath(m.MethodName, m.MethodData, m.DOMEvent)
	case EventLiveDisconnect:
		err = c.Kill()
	}

	lp.entry.UpdateWithSource(source)

	return err
}

const PageComponentUpdated = 1
const PageComponentMounted = 2

func (lp *Page) enableComponentLifeCycleReceiver() {

	go func() {
		for {
			ls := <-*lp.ComponentsLifeCycle

			switch ls.Stage {
			case Created:
				lp.Emit(PageComponentMounted, ls.Component)
				break
			case WillMount:
				break
			case Mounted:
				break
			case Updated:
				lp.EmitWithSource(PageComponentUpdated, ls.Component, ls.Source)
				break
			case WillUnmount:
				break
			case Unmounted:
				break
			case Rendered:
				break
			}
		}
	}()
}
