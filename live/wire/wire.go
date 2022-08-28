package wire

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
)

type EventSourceType string
type EventKind string

const (
	FromBrowserLiveInput         EventKind = "li"
	FromBrowserLiveMethod        EventKind = "lm"
	FromBrowserLiveDisconnect    EventKind = "lx"
	FromServerLiveDom            EventKind = "ld"
	FromServerLiveError          EventKind = "le"
	FromServerLiveConnectElement EventKind = "lce"
	FromServerLiveNavigate       EventKind = "ln"
)

type Event struct {
	OriginComponentID string `json:"component_id,omitempty"`

	Type    EventSourceType `json:"type,omitempty"`
	Kind    EventKind       `json:"kind,omitempty"`
	Value   string          `json:"t,omitempty"`
	KeyCode string          `json:"keyCode,omitempty"`

	MethodName string            `json:"method_name,omitempty"`
	MethodData map[string]string `json:"method_data,omitempty"`
	StateKey   string            `json:"key,omitempty"`
	StateValue string            `json:"value,omitempty"`

	Patches *[]PatchInstruction `json:"i,omitempty"`
}

func (e *Event) AddInstruction(instruction PatchInstruction) {
	if e.Patches == nil {
		var a []PatchInstruction
		e.Patches = &a
	}

	*e.Patches = append(*e.Patches, instruction)
}

type PatchInstruction struct {
	Name     string      `json:"n"`
	Type     string      `json:"t"`
	Attr     interface{} `json:"a,omitempty"`
	Content  string      `json:"c,omitempty"`
	Selector string      `json:"s"`
	Index    int         `json:"i,omitempty"`
}

// Wire should be responsible to keep browser view state
// equal to server view state.
// With that, it could't be responsible for the sessions.
// But the wire could belong to a session.
type Wire struct {
	ToBrowser chan Event
	log       golive.Log
	root      *component.Component
}

func NewWire() Wire {
	return Wire{
		ToBrowser: make(chan Event),
		root:      nil,
	}
}

func (w *Wire) sendToBrowser(tb Event) {
	go func() {
		w.ToBrowser <- tb
	}()
}

// Start wire up the component and all its children
func (w *Wire) Start(c *component.Component) error {
	w.root = c
	component.OnMounted(c, func(ctx *component.Context) {
		w.sendToBrowser(Event{
			OriginComponentID: ctx.Component.Name,
			Kind:              FromServerLiveConnectElement,
		})
	})

	component.OnUpdate(c, func(ctx *component.Context) {

		c.Log(golive.LogDebug, "Wire: OnUpdate", golive.LogEx{"name": c.Name, "ctx_closed": ctx.Closed, "ctx_frozen": ctx.Frozen})
		w.LiveRenderComponent(ctx.Component, &Event{})
	})

	component.Provide(c, "wire", w)
	err := c.Mount()
	if err != nil {
		return fmt.Errorf("mount wire component: %w", err)
	}

	return nil
}

func (w *Wire) HandleFromBrowser(e *Event) {

	var err error

	c, err := w.root.FindComponent(e.OriginComponentID)
	if c == nil {
		w.log(golive.LogError, fmt.Sprintf("handle browser event: %s", err), golive.LogEx{})
		return
	}

	switch e.Kind {
	case FromBrowserLiveInput:
		err = c.State.SetValueInPath(e.StateValue, e.StateKey)
	case FromBrowserLiveMethod:
		_, err = c.State.InvokeMethodInPath(e.MethodName, []reflect.Value{reflect.ValueOf(e.MethodData), reflect.ValueOf(e.Kind)})
	case FromBrowserLiveDisconnect:
		err = c.Unmount()
	}

	if err != nil {
		w.log(golive.LogError, fmt.Sprintf("handle browser event: %s", err), golive.LogEx{})
	}

	// TODO: find some way to call update passing event avoiding multiple transfers
	c.Update()
}

func (w *Wire) NavigateToPage(path string) {
	w.sendToBrowser(Event{
		OriginComponentID: w.root.Name,
		Kind:              FromServerLiveNavigate,
		Value:             path,
	})
}

// LiveRenderComponent render component and detect diff
// from the last component render state convert the diff
// to patches to send to browser
func (w *Wire) LiveRenderComponent(c *component.Component, e *Event) error {
	var err error

	c.Log(golive.LogDebug, "LiveRenderComponent", golive.LogEx{"name": c.Name})

	from := c.RenderState.GetHTML()

	err = c.Render()
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	to := c.RenderState.GetHTML()
	d := differ.NewDiff(from)
	d.Propose(to)

	if err != nil {
		return fmt.Errorf("live render: %w", err)
	}

	patches, err := diffToBrowser(d, e)

	if err != nil {
		return fmt.Errorf("diff to browser: %w", err)
	}

	for _, patch := range patches {
		w.sendToBrowser(*patch)
	}

	return nil
}

func (w *Wire) End() error {
	return w.root.Unmount()
}

func diffToBrowser(diff *differ.Diff, source *Event) ([]*Event, error) {

	bp := make([]*Event, 0)

	for _, instruction := range diff.Instructions {

		selector, err := dom.SelectorFromNode(instruction.Element)
		if isUpdateAbleToSkip(instruction, source) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("selector from node: %w instruction: %v", err, instruction)
		}

		componentID, err := renderer.ComponentIDFromNode(instruction.Element)

		if err != nil {
			return nil, fmt.Errorf("get component id from node: %w", err)
		}

		var tb *Event

		// find if there is already a tb
		for _, pb := range bp {
			if pb.OriginComponentID == componentID {
				tb = pb
				break
			}
		}

		// If there is no tb
		if tb == nil {
			tb = &Event{OriginComponentID: componentID}
			tb.Kind = FromServerLiveDom
			bp = append(bp, tb)
		}

		tb.AddInstruction(PatchInstruction{
			Type: instruction.ChangeType.ToString(),
			Attr: map[string]string{
				"Name":  instruction.Attr.Name,
				"Value": instruction.Attr.Value,
			},
			Index:    instruction.Index,
			Content:  instruction.Content,
			Selector: selector.ToString(),
		})
	}
	return bp, nil
}

func isUpdateAbleToSkip(in differ.ChangeInstruction, event *Event) bool {
	if in.Element == nil || event == nil || in.ChangeType != differ.SetAttr || strings.ToLower(in.Attr.Name) != "value" {
		return false
	}

	attr := dom.GetAttribute(in.Element, component.GoLiveInput)

	return attr != nil && attr.Val == event.Value
}
