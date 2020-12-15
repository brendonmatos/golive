package golive

const (
	EventLiveInput      = "li"
	EventLiveMethod     = "lm"
	EventLiveDom        = "ld"
	EventLiveDisconnect = "lx"
)

type InMessage struct {
	Name         string `json:"name"`
	ScopeID      string `json:"scope_id"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
	StateKey     string `json:"key"`
	StateValue   string `json:"value"`
}

type OutMessage struct {
	Name    string      `json:"name"`
	ScopeID string      `json:"scope_id"`
	Type    string      `json:"type"`
	Attr    interface{} `json:"attr"`
	Content string      `json:"content"`
	Element string      `json:"element"`
}
