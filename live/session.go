package live

import (
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/wire"
)

var (
	ErrorSessionNotFound = "session_not_found"
)

func ErrorMap() map[string]string {
	return map[string]string{
		"LiveErrorSessionNotFound": ErrorSessionNotFound,
	}
}

type SessionStatus string

const (
	SessionNew    SessionStatus = "n"
	SessionOpen   SessionStatus = "o"
	SessionClosed SessionStatus = "c"
)

type Session struct {
	log    golive.Log
	Status SessionStatus
	Wire   *wire.Wire
}

func (s *Session) WireComponent(lc *component.Component) error {
	s.Wire = wire.NewWire(lc)
	err := s.Wire.Start()
	if err != nil {
		return fmt.Errorf("wire start: %w", err)
	}
	return nil
}

func (s *Session) SetOpen() {
	s.Status = SessionOpen
}

func (s *Session) SetClosed() {
	s.Status = SessionClosed
}

func NewSession() *Session {
	return &Session{
		Wire:   nil,
		Status: SessionNew,
	}
}
