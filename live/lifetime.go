package live

type LifeTimeStage int

const (
	WillCreate LifeTimeStage = iota
	Created

	WillMount

	WillMountChildren
	ChildrenMounted

	Mounted

	Rendered
	Updated

	WillUnmount
	Unmounted
)

type LifeTimeMessage struct {
	Stage     LifeTimeStage
	Component *Component
	Source    *EventSource
}

type LifeCycle chan LifeTimeMessage

type EventSource struct {
	Type  EventSourceType
	Value string
}

type EventSourceType string

const (
	EventSourceInput = "input"
)
