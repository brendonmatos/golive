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
	Session          SessionKey
	Events           LiveEventsChannel
	Entry            *LiveComponent
	ComponentUpdates *LifeTimeUpdates
	Components       map[string]*LiveComponent
	RenderedContent  string
}

type PageContent struct {
	Lang   string
	Body   template.HTML
	Script string
	Title  string
	Enum   PageEnum
}

func NewLivePage(s SessionKey, c *LiveComponent) *LivePage {
	componentsUpdatesChannel := make(LifeTimeUpdates)
	pageEventsChannel := make(LiveEventsChannel)

	c.UpdatesChannel = &componentsUpdatesChannel

	return &LivePage{
		Session:          s,
		Entry:            c,
		Events:           pageEventsChannel,
		ComponentUpdates: &componentsUpdatesChannel,
		Components:       make(map[string]*LiveComponent),
	}
}

func (lp *LivePage) handleComponentsLifeTime() {

	go func() {
		for {
			update := <-*lp.ComponentUpdates

			switch update.Stage {
			case WillMount:
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
			case Unmounted:
				lp.Components[update.Component.Name] = nil
				break

			case Rendered:
				break
			}
		}
	}()
}

func (lp *LivePage) Prepare() {
	lp.handleComponentsLifeTime()
	lp.Entry.Prepare()
}

func (lp *LivePage) Mount() {
	lp.Entry.Mount()
}

func (lp *LivePage) FirstRender(pc PageContent) (string, error) {
	rendered := lp.Entry.Render()

	writer := bytes.NewBufferString("")

	pc.Body = template.HTML(rendered)
	pc.Enum = PageEnum{
		EventLiveDom:    EventLiveDom,
		EventLiveInput:  EventLiveInput,
		EventLiveMethod: EventLiveMethod,
	}

	err := BasePage.Execute(writer, pc)

	if err != nil {
		return "", err
	}

	return writer.String(), nil
}

func (lp *LivePage) HandleMessage(m InMessage) error {

	c, ok := lp.Components[m.ScopeID]

	if !ok {
		return fmt.Errorf("c not found with id: %s", m.ScopeID)
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
