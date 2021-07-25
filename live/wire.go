package live

import (
	"github.com/brendonmatos/golive/live/util"
)

// Wire should be responsible to keep browser view state
// equal to server view state.
// With that, it could not be responsible for the sessions.
// But the wire should belong to a session. TODO
type Wire struct {
	Sessions Sessions
}

type Sessions map[string]*Session

func NewWire() *Wire {
	return &Wire{
		// TODO: move sessions to server
		Sessions: make(Sessions),
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
