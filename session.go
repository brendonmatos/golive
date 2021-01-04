package golive

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
	s.LivePage.ForceUpdate()
	return nil
}

func (s *Session) ActivatePage(lp *Page) {
	s.LivePage = lp

	// Pre-render to ensure we have something to diff against
	for _, component := range lp.Components {
		if component.rendered == "" {
			if err := s.LiveRenderComponent(component); err != nil {
				s.log(LogError, "activate page", logEx{"error": err})
			}
		}
	}

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from the page!
			pageUpdate := <-lp.Events
			if pageUpdate.Type == Updated {
				if err := s.LiveRenderComponent(pageUpdate.Component); err != nil {
					s.log(LogError, "component live render", logEx{"error": err})
				}
			}
			if pageUpdate.Type == Unmounted {
				return
			}
		}
	}()
}

func (s *Session) LiveRenderComponent(c *LiveComponent) error {
	var err error

	changes, err := c.LiveRender()

	if err != nil {
		return err
	}

	s.QueueMessage(*changes)

	return nil
}
