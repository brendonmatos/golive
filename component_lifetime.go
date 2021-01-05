package golive

type LifeTimeStage int

const (
	WillMount LifeTimeStage = iota
	Mounted
	Rendered
	Updated
	WillUnmount
	Unmounted

	WillMountChildren
	ChildrenMounted

	WillPrepareChildren
	ChildrenPrepared
)

type ComponentLifeTimeMessage struct {
	Stage     LifeTimeStage
	Component *LiveComponent
}

type ComponentLifeCycle chan ComponentLifeTimeMessage
