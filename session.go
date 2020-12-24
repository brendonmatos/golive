package golive

const (
	EventLiveInput      = "li"
	EventLiveMethod     = "lm"
	EventLiveDom        = "ld"
	EventLiveDisconnect = "lx"
)

type InMessage struct {
	Name         string `json:"name"`
	ComponentId  string `json:"component_id"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
	StateKey     string `json:"key"`
	StateValue   string `json:"value"`
}

type OutMessage struct {
	Name        string      `json:"name"`
	ComponentId string      `json:"component_id"`
	Type        string      `json:"type"`
	Attr        interface{} `json:"attr,omitempty"`
	Content     string      `json:"content,omitempty"`
	Element     string      `json:"element"`
}

type Session struct {
	LivePage   *Page
	OutChannel chan OutMessage
}

func NewSession() *Session {
	return &Session{
		OutChannel: make(chan OutMessage),
	}
}

func (s *Session) QueueMessage(message OutMessage) {
	go func() {
		s.OutChannel <- message
	}()
}

func (s *Session) QueueMessages(messages []OutMessage) {
	for _, message := range messages {
		s.QueueMessage(message)
	}
}

func (s *Session) IngestMessage(message InMessage) error {
	err := s.LivePage.HandleMessage(message)
	if err != nil {
		return err
	}
	s.LivePage.ForceUpdate()
	return nil
}

func (s *Session) ActivatePage(lp *Page) {
	s.LivePage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from the page!
			pageUpdate := <-lp.Events
			if pageUpdate.Type == Updated {
				_ = s.LiveRenderComponent(pageUpdate.Component)
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

	s.QueueMessages(changes)

	return nil
}
