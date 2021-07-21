package live

import "github.com/brendonmatos/golive/live/component"

type LifeTimeMessage struct {
	Stage     string
	Component *component.Component
	Source    *EventSource
}

type EventSource struct {
	Type  EventSourceType
	Value string
}

type EventSourceType string

const (
	EventSourceInput = "input"
)
