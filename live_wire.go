package golive

import "fmt"

type WireMessage struct {
	SessionKey SessionKey
	Broadcast  bool
	OutMessage OutMessage
}

type WiredSession struct {
	LivePage    *LivePage
	OutMessages chan OutMessage
}

type WiredSessions map[SessionKey]*WiredSession

type LiveWire struct {
	Sessions WiredSessions
}

func NewWire() *LiveWire {
	return &LiveWire{

		// Sessions from all pages!
		Sessions: make(WiredSessions),
	}
}

func (w *LiveWire) CreateSession() (SessionKey, error) {
	session := NewSessionKey()

	w.Sessions[session] = &WiredSession{
		LivePage:    nil,
		OutMessages: make(chan OutMessage),
	}

	return session, nil
}

func (w *LiveWire) QueueMessage(sKey SessionKey, message OutMessage) error {
	oc, err := w.GetOutChannel(sKey)

	if err != nil {
		return err
	}

	go func() {
		*oc <- message
	}()

	return nil
}

func (w *LiveWire) IngestMessage(session SessionKey, message InMessage) error {
	s := w.Sessions[session]

	if s == nil {
		return fmt.Errorf("session without any page associated, key=%v", session)
	}

	err := s.LivePage.HandleMessage(message)

	if err != nil {
		return err
	}

	s.LivePage.ForceUpdate()

	return nil
}

func (w *LiveWire) updateWiredLiveComponent(session SessionKey, c *LiveComponent) error {
	var err error

	_, changes := c.LiveRender()

	for _, change := range changes {
		err = w.QueueMessage(session, change)
	}

	return err
}

func (w *LiveWire) ActivateLivePage(session SessionKey, lp *LivePage) {
	w.Sessions[session].LivePage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from the page!
			pageUpdate := <-lp.Events
			if pageUpdate.Type == Updated {
				_ = w.updateWiredLiveComponent(session, pageUpdate.Component)
			}
			if pageUpdate.Type == Unmounted {
				return
			}
		}
	}()
}

func (w *LiveWire) GetOutChannel(key SessionKey) (*chan OutMessage, error) {
	s := w.Sessions[key]

	if s == nil || s.OutMessages == nil || &s.OutMessages == nil {
		return nil, fmt.Errorf("session is no wired correcly, key=%v", key)
	}

	return &s.OutMessages, nil
}
