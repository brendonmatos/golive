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
}

type ComponentLifeCycle chan ComponentLifeTimeMessage
