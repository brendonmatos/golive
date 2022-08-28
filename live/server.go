package live

import (
	"errors"
	"fmt"

	"github.com/brendonmatos/golive/live/component"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) CreateSession() *Session {
	/* Create session to the new user */
	session := NewSession()
	return session
}

func (s *Server) CreateLivePage(session *Session, lc *component.Component, pc PageContent) (*string, error) {

	// Instantiate a page to attach to a session
	p := NewLivePage()

	// Set page content
	p.SetContent(pc)
	p.SetRootComponent(lc)

	// activation should be before mount,
	// because in activation will setup page channels
	// that will be needed in mount
	err := session.WireComponent(lc)
	if err != nil {
		return nil, fmt.Errorf("session wire component: %w", err)
	}

	rendered, err := p.Render()

	return &rendered, err
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
