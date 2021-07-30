package wire

type Event struct {
	Type    EventSourceType
	Value   string
	KeyCode string `json:"keyCode"`
}

type EventSourceType string

const (
	EventSourceInput = "input"
)
