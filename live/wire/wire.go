package wire

import (
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
	"reflect"
	"strings"
)

const (
	FromBrowserLiveInput        = "li"
	FromBrowserLiveMethod       = "lm"
	FromBrowserLiveDisconnect   = "lx"
	ToBrowserLiveDom            = "ld"
	ToBrowserLiveError          = "le"
	ToBrowserLiveConnectElement = "lce"
	ToBrowserLiveNavigate       = "ln"
)

// Wire should be responsible to keep browser view state
// equal to server view state.
// With that, it could not be responsible for the sessions.
// But the wire should belong to a session. TODO
type Wire struct {
	ToBrowser chan ToBrowser
	log       golive.Log
	root      *component.Component
}

func NewWire(root *component.Component) *Wire {
	return &Wire{
		ToBrowser: make(chan ToBrowser),
		root:      root,
	}
}

type ToBrowser struct {
	Type         string `json:"t"`
	ComponentID  string `json:"cid,omitempty"`
	Event        *Event
	Value        string              `json:"value"`
	Message      string              `json:"m"`
	Instructions *[]PatchInstruction `json:"i,omitempty"`
}

func (b *ToBrowser) AddInstruction(instruction PatchInstruction) {
	if b.Instructions == nil {
		var a []PatchInstruction
		b.Instructions = &a
	}

	*b.Instructions = append(*b.Instructions, instruction)
}

type FromBrowser struct {
	Name        string            `json:"name"`
	ComponentID string            `json:"component_id"`
	MethodName  string            `json:"method_name"`
	MethodData  map[string]string `json:"method_data"`
	StateKey    string            `json:"key"`
	StateValue  string            `json:"value"`
	Event       *Event            `json:"dom_event"`
}

func (w *Wire) sendToBrowser(tb ToBrowser) {
	go func() {
		w.ToBrowser <- tb
	}()
}

// Start wire up the component and all its children
func (w *Wire) Start() error {
	c := w.root
	component.OnMounted(c, func(ctx *component.Context) {
		w.sendToBrowser(ToBrowser{
			Type:        ToBrowserLiveConnectElement,
			ComponentID: ctx.Component.Name,
			Event:       nil,
		})
	})

	component.OnUpdate(c, func(ctx *component.Context) {
		w.LiveRenderComponent(ctx.Component, &Event{})
	})

	component.Provide(c, "wire", w)
	err := c.Mount()
	if err != nil {
		return fmt.Errorf("mount wire component: %w", err)
	}

	return nil
}

func (w *Wire) HandleFromBrowser(m *FromBrowser) {

	var err error

	c, err := w.root.FindComponent(m.ComponentID)
	if c == nil {
		w.log(golive.LogError, fmt.Sprintf("handle browser event: %s", err), golive.LogEx{})
		return
	}

	switch m.Name {
	case FromBrowserLiveInput:
		err = c.State.SetValueInPath(m.StateValue, m.StateKey)
	case FromBrowserLiveMethod:
		_, err = c.State.InvokeMethodInPath(m.MethodName, []reflect.Value{reflect.ValueOf(m.MethodData), reflect.ValueOf(m.Event)})
	case FromBrowserLiveDisconnect:
		err = c.Unmount()
	}

	// TODO: find some way to call update passing event avoiding multiple transfers
	c.Update()
}

// LiveRender call it when you need to force update
func (w *Wire) LiveRender() error {
	err := w.LiveRenderComponent(w.root, &Event{})
	if err != nil {
		return fmt.Errorf("live render component: %w", err)
	}
	return nil
}

func (w *Wire) NavigateToPage(path string) {
	w.sendToBrowser(ToBrowser{
		ComponentID: w.root.Name,
		Type:        ToBrowserLiveNavigate,
		Value:       path,
	})
}

func (w *Wire) LiveConnectElement(c *component.Component) {
	w.sendToBrowser(ToBrowser{
		ComponentID: c.Name,
		Type:        ToBrowserLiveConnectElement,
	})
}

// LiveRenderComponent render the updated Component and compare with
// last state. It may apply with *all child componentsRegister*
func (w *Wire) LiveRenderComponent(c *component.Component, e *Event) error {
	var err error

	diff, err := c.LiveRender()

	if err != nil {
		return err
	}

	patches, err := diffToBrowser(diff, e)

	if err != nil {
		return err
	}

	for _, patch := range patches {
		w.sendToBrowser(*patch)
	}

	return nil
}

func (w *Wire) End() error {
	return w.root.Unmount()
}

func diffToBrowser(diff *differ.Diff, source *Event) ([]*ToBrowser, error) {

	bp := make([]*ToBrowser, 0)

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

		var tb *ToBrowser

		// find if there is already a tb
		for _, pb := range bp {
			if pb.ComponentID == componentID {
				tb = pb
				break
			}
		}

		// If there is no tb
		if tb == nil {
			tb = &ToBrowser{ComponentID: componentID}
			tb.Type = ToBrowserLiveDom
			bp = append(bp, tb)
		}

		tb.AddInstruction(PatchInstruction{
			Name: ToBrowserLiveDom,
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

	return attr != nil && event.Type == FromBrowserLiveInput && attr.Val == event.Value
}
