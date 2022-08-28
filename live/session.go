package live

import (
	"fmt"

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
	SessionIdle   SessionStatus = "i"
	SessionOpen   SessionStatus = "o"
	SessionClosed SessionStatus = "c"
)

type Session struct {
	Status     SessionStatus
	Wire       *wire.Wire
	ExitSignal chan bool
}

func (s *Session) WireComponent(lc *component.Component) error {
	s.Wire = wire.NewWire(lc)
	err := s.Wire.Start()
	if err != nil {
		return fmt.Errorf("wire start: %w", err)
	}
	return nil
}

func (s *Session) Close() {
	s.Status = SessionClosed
	err := s.Wire.End()
	if err != nil {
		panic(err)
	}
	s.ExitSignal <- true
	close(s.ExitSignal)
}

func (s *Session) IsClosed() bool {
	return s.Status == SessionClosed
}

func (s *Session) IsOpen() bool {
	return s.Status == SessionOpen
}

func NewSession() *Session {
	return &Session{
		Wire:   nil,
		Status: SessionOpen,
	}
}
