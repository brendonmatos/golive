package golive

import "fmt"

type WireMessage struct {
	SessionKey SessionKey
	OutMessage OutMessage
}

type WiredSessions map[SessionKey]*LivePage

type LiveWire struct {
	Sessions   WiredSessions
	OutChannel chan WireMessage
}

func NewWire() *LiveWire {
	return &LiveWire{

		// Sessions from all pages!
		Sessions: make(map[SessionKey]*LivePage),

		// OutChannel to all the sockets!
		OutChannel: make(chan WireMessage),
	}
}

func (w *LiveWire) CreateSession() (SessionKey, error) {
	session := NewSessionKey()
	return session, nil
}

func (w *LiveWire) QueueMessage(sKey SessionKey, message OutMessage) {
	w.OutChannel <- WireMessage{
		SessionKey: sKey,
		OutMessage: message,
	}
}

func (w *LiveWire) HandleMessage(session SessionKey, message InMessage) error {
	pg := w.Sessions[session]

	if pg == nil {
		return fmt.Errorf("session without any page associated, key=%v", session)
	}

	err := pg.HandleMessage(message)

	if err != nil {
		return err
	}

	pg.Events <- LivePageEvent{
		Type:      Updated,
		Component: pg.Entry,
	}

	return nil
}

func (w *LiveWire) NotifyPageChanges(session SessionKey, c *LiveComponent) error {
	_, changes := c.LiveRender()

	for _, change := range changes {
		w.QueueMessage(session, change)
	}

	return nil
}

func (w *LiveWire) SetSession(session SessionKey, lp *LivePage) {
	w.Sessions[session] = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			pageUpdate := <-lp.Events
			if pageUpdate.Type == Updated {
				_ = w.NotifyPageChanges(session, pageUpdate.Component)
			}
		}
	}()
}
