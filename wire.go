package golive

type LiveWire struct {
	Sessions WireSessions
}

type WireSessions map[string]*Session

func NewWire() *LiveWire {
	return &LiveWire{
		Sessions: make(WireSessions),
	}
}

func (w *LiveWire) GetSession(s string) *Session {
	return w.Sessions[s]
}
func (w *LiveWire) DeleteSession(s string) {
	delete(w.Sessions, s)
}

func (w *LiveWire) CreateSession() (string, *Session, error) {
	key, _ := GenerateRandomString(48)
	s := NewSession()
	w.Sessions[key] = s
	return key, s, nil
}
