package golive

const (
	WillMount = iota
	Mounted
	Rendered
	Updated
	WillUnmount
	Unmounted
)

type ComponentLifeTimeMessage struct {
	Stage     int
	Component *LiveComponent
}

type LifeTimeUpdates chan ComponentLifeTimeMessage
