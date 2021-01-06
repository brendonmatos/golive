package golive

import "strconv"

const (
	EventLiveInput      = "li"
	EventLiveMethod     = "lm"
	EventLiveDom        = "ld"
	EventLiveDisconnect = "lx"
)

type BrowserEvent struct {
	Name         string `json:"name"`
	ComponentID  string `json:"component_id"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
	StateKey     string `json:"key"`
	StateValue   string `json:"value"`
}

type Session struct {
	LivePage   *Page
	OutChannel chan PatchBrowser
	log        Log
}

func NewSession() *Session {
	return &Session{
		OutChannel: make(chan PatchBrowser),
	}
}

func (s *Session) QueueMessage(message PatchBrowser) {
	go func() {
		s.OutChannel <- message
	}()
}

func (s *Session) IngestMessage(message BrowserEvent) error {

	err := s.LivePage.HandleMessage(message)

	if err != nil {
		return err
	}

	s.LivePage.SendUpdate()

	return nil
}

func (s *Session) ActivatePage(lp *Page) {
	s.LivePage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from page
			evt := <-lp.SessionEvents
			if evt.Type == int(Updated) {
				if err := s.LiveRenderComponent(evt.Component); err != nil {
					s.log(LogError, "component live render", logEx{"error": err})
				}
			}
			if evt.Type == int(Unmounted) {
				// TODO: Treat unmount
				return
			}
		}
	}()
}

func (s *Session) LiveRenderComponent(c *LiveComponent) error {
	var err error

	diff, err := c.LiveRender()

	if err != nil {
		return err
	}

	om := NewPatchBrowser(c.Name)
	om.Name = EventLiveDom

	for _, instruction := range diff.instructions {

		selector, err := SelectorFromNode(instruction.Element)

		if err != nil {
			return err
		}

		om.AddInstruction(PatchInstruction{
			Name:     EventLiveDom,
			Type:     strconv.Itoa(int(instruction.Type)),
			Attr:     instruction.Attr,
			Content:  instruction.Content,
			Selector: selector,
		})
	}

	s.QueueMessage(*om)
	return nil
}
