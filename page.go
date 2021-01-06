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
	EventLiveInput   string
	EventLiveMethod  string
	EventLiveDom     string
	DiffSetAttr      string
	DiffRemoveAttr   string
	DiffReplace      string
	DiffRemove       string
	DiffSetInnerHtml string
	DiffAppend       string
}

type LivePageEvent struct {
	Type      int
	Component *LiveComponent
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
	Lang   string
	Body   template.HTML
	Head   template.HTML
	Script string
	Title  string
	Enum   PageEnum
}

func NewLivePage(c *LiveComponent) *Page {
	componentsUpdatesChannel := make(ComponentLifeCycle)
	pageEventsChannel := make(LiveEventsChannel)

	c.updatesChannel = &componentsUpdatesChannel

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

func (lp *Page) Prepare() {
	lp.handleComponentsLifeTime()
	lp.entry.Prepare()
}

func (lp *Page) Mount() {
	lp.entry.Mount()
}

func (lp *Page) Render() (string, error) {
	rendered, _ := lp.entry.Render()

	lp.content.Body = template.HTML(rendered)
	lp.content.Enum = PageEnum{
		EventLiveInput:   EventLiveInput,
		EventLiveMethod:  EventLiveMethod,
		EventLiveDom:     EventLiveDom,
		DiffSetAttr:      SetAttr.String(),
		DiffRemoveAttr:   RemoveAttr.String(),
		DiffReplace:      Replace.String(),
		DiffRemove:       Remove.String(),
		DiffSetInnerHtml: SetInnerHtml.String(),
		DiffAppend:       Append.String(),
	}

	writer := bytes.NewBufferString("")
	err := BasePage.Execute(writer, lp.content)
	return writer.String(), err
}

func (lp *Page) ForceUpdate() {
	lp.Events <- LivePageEvent{
		Type:      Updated,
		Component: lp.entry,
	}
}

func (lp *Page) HandleMessage(m InMessage) error {

	c, ok := lp.Components[m.ComponentId]

	if !ok {
		return fmt.Errorf("component not found with id: %s", m.ComponentId)
	}

	switch m.Name {
	case EventLiveInput:
		{
			return c.SetValueInPath(m.StateValue, m.StateKey)
		}
	case EventLiveMethod:
		{
			return c.InvokeMethodInPath(m.MethodName, m.MethodData, m.DOMEvent)
		}
	case EventLiveDisconnect:
		{
			return c.Kill()
		}
	}

	return nil
}

func (lp *Page) handleComponentsLifeTime() {

	go func() {
		for {
			update := <-*lp.ComponentsLifeCycle

			switch update.Stage {
			case WillMount:
				break
			case Mounted:
				lp.Components[update.Component.Name] = update.Component
				break
			case Updated:
				lp.Events <- LivePageEvent{
					Type:      Updated,
					Component: update.Component,
				}
				break
			case WillUnmount:
				break
			case Unmounted:
				lp.Components[update.Component.Name] = nil
				break
			case Rendered:
				break
			}
		}
	}()
}
