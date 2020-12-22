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
	EventLiveInput  string
	EventLiveMethod string
	EventLiveDom    string
}

type LivePageEvent struct {
	Type      int
	Component *LiveComponent
}

type LiveEventsChannel chan LivePageEvent

type LivePage struct {
	Session             SessionKey
	Events              LiveEventsChannel
	ComponentsLifeCycle *ComponentLifeCycle

	entry *LiveComponent

	// Components is a list that handle all the components from the page
	Components map[string]*LiveComponent
}

type PageContent struct {
	Lang   string
	Body   template.HTML
	Script string
	Title  string
	Enum   PageEnum
}

func NewLivePage(s SessionKey, c *LiveComponent) *LivePage {
	componentsUpdatesChannel := make(ComponentLifeCycle)
	pageEventsChannel := make(LiveEventsChannel)

	c.updatesChannel = &componentsUpdatesChannel

	return &LivePage{
		entry:               c,
		Events:              pageEventsChannel,
		ComponentsLifeCycle: &componentsUpdatesChannel,
		Components:          make(map[string]*LiveComponent),
	}
}

func (lp *LivePage) Prepare() {
	lp.handleComponentsLifeTime()
	lp.entry.Prepare()
}

func (lp *LivePage) Mount() {
	lp.entry.Mount()
}

func (lp *LivePage) FirstRender(pc PageContent) (string, error) {
	rendered := lp.entry.Render()

	pc.Body = template.HTML(rendered)
	pc.Enum = PageEnum{
		EventLiveDom:    EventLiveDom,
		EventLiveInput:  EventLiveInput,
		EventLiveMethod: EventLiveMethod,
	}

	writer := bytes.NewBufferString("")
	err := BasePage.Execute(writer, pc)
	return writer.String(), err
}

func (lp *LivePage) ForceUpdate() {
	lp.Events <- LivePageEvent{
		Type:      Updated,
		Component: lp.entry,
	}
}

func (lp *LivePage) HandleMessage(m InMessage) error {

	c, ok := lp.Components[m.ScopeID]

	if !ok {
		return fmt.Errorf("component not found with id: %s", m.ScopeID)
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

func (lp *LivePage) handleComponentsLifeTime() {

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
