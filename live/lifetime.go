package live

type LifeTimeMessage struct {
	Stage     string
	Component *Component
	Source    *EventSource
}

//
//type LifeCycle chan LifeTimeMessage
//
type EventSource struct {
	Type  EventSourceType
	Value string
}

type EventSourceType string

const (
	EventSourceInput = "input"
)
