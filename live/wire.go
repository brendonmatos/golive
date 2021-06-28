package live

import (
	"github.com/brendonmatos/golive/live/util"
)

type Wire struct {
	Sessions WireSessions
}

type WireSessions map[string]*Session

func NewWire() *Wire {
	return &Wire{
		Sessions: make(WireSessions),
	}
}

func (w *Wire) GetSession(s string) *Session {
	return w.Sessions[s]
}
func (w *Wire) DeleteSession(s string) {
	delete(w.Sessions, s)
}

func (w *Wire) CreateSession() (string, *Session, error) {
	key, _ := util.GenerateRandomString(48)
	s := NewSession()
	w.Sessions[key] = s
	return key, s, nil
}
