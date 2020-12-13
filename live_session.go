package golive

import "github.com/google/uuid"

type SessionKey string

func NewSessionKey() SessionKey {
	newUUID, _ := uuid.NewUUID()

	return SessionKey(newUUID.String())
}
