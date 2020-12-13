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
		Sessions:   make(map[SessionKey]*LivePage),
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

	// TODO: Verify if this is the right place to call the updates from live base component
	component, _ := pg.Component.FindComponent(message.ScopeID)

	switch message.Name {
	case EventLiveInput:
		{
			component.SetValueInPath(message.StateValue, message.StateKey)
		}
	case EventLiveMethod:
		{
			component.InvokeMethodInPath(message.MethodName, message.MethodParams)
		}
	case EventLiveDisconnect:
		{
			_ = component.Kill()
		}
	}

	_ = w.SendComponentsChange(session, component)

	return nil
}

func (w *LiveWire) SendComponentsChange(session SessionKey, c *LiveComponent) error {
	_, changes := c.LiveRender()

	for _, change := range changes {
		w.QueueMessage(session, change)
	}

	return nil
}

func (w *LiveWire) SetSession(session SessionKey, lp *LivePage) {
	w.Sessions[session] = lp

	go func() {
		for {

			commit := <-*lp.ComponentsLifeTimeChannel

			switch commit {
			case LifeTimeUpdate:
				_ = w.SendComponentsChange(session, lp.Component)
			case LifeTimeExit:
				return
			}

		}
	}()
}
