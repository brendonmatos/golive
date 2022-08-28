package live

import (
	"errors"
	"fmt"

	"github.com/brendonmatos/golive/live/component"
)

type Server struct {
	OnSession func(*Session)
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) CreatePageAccess(lc *component.Component, c PageContent, session *Session) (*string, *Session, error) {

	/* Create session to the new user */
	if session == nil {
		session = NewSession()
	}

	// Instantiate a page to attach to a session
	p := NewLivePage(lc)

	// Set page content
	p.SetContent(c)

	// activation should be before mount,
	// because in activation will setup page channels
	// that will be needed in mount
	err := session.WireComponent(lc)
	if err != nil {
		return nil, nil, fmt.Errorf("session wire component: %w", err)
	}

	rendered, err := p.Render()

	return &rendered, session, err
}

func (s *Server) HandleSessionRealtime(session *Session) error {
	defer func() {
		payload := recover()
		if payload != nil {
			// s.Log(golive.LogWarn, fmt.Sprintf("ws request panic recovered: %v", payload), nil)
		}
	}()

	if session == nil {
		return errors.New("session is not right")
	}

	return nil
}
