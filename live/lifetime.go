package live

type LifeTimeMessage struct {
	Stage     string
	Component *Component
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
