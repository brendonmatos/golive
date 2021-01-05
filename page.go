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
	DiffSetAttr      DiffType
	DiffRemoveAttr   DiffType
	DiffReplace      DiffType
	DiffRemove       DiffType
	DiffSetInnerHTML DiffType
	DiffAppend       DiffType
}

type LivePageEvent struct {
	Type      int
	Component *LiveComponent
}

type LiveEventsChannel chan LivePageEvent

type Page struct {
	content             PageContent
	SessionEvents       LiveEventsChannel
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

	return &Page{
		entry:               c,
		SessionEvents:       pageEventsChannel,
		ComponentsLifeCycle: &componentsUpdatesChannel,
		Components:          make(map[string]*LiveComponent),
	}
}

func (lp *Page) SetContent(c PageContent) {
	lp.content = c
}

func (lp *Page) Mount() {
	lp.handleComponentsLifeTime()

	// Call the component in sequence of life cycle
	err := lp.entry.Create()

	if err != nil {
		panic(err)
	}

	lp.entry.Prepare(lp.ComponentsLifeCycle)
	lp.entry.Mount()
}

func (lp *Page) Render() (string, error) {
	rendered, err := lp.entry.Render()

	if err != nil {
		return "", err
	}

	lp.entry.rendered = rendered

	lp.content.Body = template.HTML(rendered)
	lp.content.Enum = PageEnum{
		EventLiveInput:   EventLiveInput,
		EventLiveMethod:  EventLiveMethod,
		EventLiveDom:     EventLiveDom,
		DiffSetAttr:      SetAttr,
		DiffRemoveAttr:   RemoveAttr,
		DiffReplace:      Replace,
		DiffRemove:       Remove,
		DiffSetInnerHTML: SetInnerHtml,
		DiffAppend:       Append,
	}

	writer := bytes.NewBufferString("")
	err = BasePage.Execute(writer, lp.content)
	return writer.String(), err
}

func (lp *Page) SendSessionEvent(lts LifeTimeStage, c *LiveComponent) {
	if c == nil {
		c = lp.entry
	}

	go func() {
		lp.SessionEvents <- LivePageEvent{
			Type:      int(lts),
			Component: c,
		}
	}()
}

func (lp *Page) SendUpdate() {
	lp.SendSessionEvent(Updated, nil)
}

func (lp *Page) HandleMessage(m BrowserEvent) error {

	c, ok := lp.Components[m.ComponentID]

	if !ok {
		return fmt.Errorf("component not found with id: %s", m.ComponentID)
	}

	switch m.Name {
	case EventLiveInput:
		{
			return c.SetValueInPath(m.StateValue, m.StateKey)
		}
	case EventLiveMethod:
		{
			return c.InvokeMethodInPath(m.MethodName, m.MethodParams)
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
				lp.SendSessionEvent(Updated, update.Component)
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
