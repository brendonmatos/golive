package golive

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

type ComponentLifeTimeMessage struct {
	Stage     LifeTimeStage
	Component *LiveComponent
	Source    *EventSource
}

type ComponentLifeCycle chan ComponentLifeTimeMessage

type EventSource struct {
	Type  EventSourceType
	Value string
}

type EventSourceType string

const (
	EventSourceInput = "input"
)
