package live

import (
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"strings"
)

const (
	EventLiveInput          = "li"
	EventLiveMethod         = "lm"
	EventLiveDom            = "ld"
	EventLiveDisconnect     = "lx"
	EventLiveError          = "le"
	EventLiveConnectElement = "lce"
)

const GoLiveInput = "gl-input"

var (
	ErrorSessionNotFound = "session_not_found"
)

func ErrorMap() map[string]string {
	return map[string]string{
		"LiveErrorSessionNotFound": ErrorSessionNotFound,
	}
}

type BrowserEvent struct {
	Name        string            `json:"name"`
	ComponentID string            `json:"component_id"`
	MethodName  string            `json:"method_name"`
	MethodData  map[string]string `json:"method_data"`
	StateKey    string            `json:"key"`
	StateValue  string            `json:"value"`
	DOMEvent    *DOMEvent         `json:"dom_event"`
}

type DOMEvent struct {
	KeyCode string `json:"keyCode"`
}

type SessionStatus string

const (
	SessionNew    SessionStatus = "n"
	SessionOpen   SessionStatus = "o"
	SessionClosed SessionStatus = "c"
)

type Session struct {
	LivePage   *Page
	OutChannel chan differ.PatchBrowser
	log        golive.Log
	Status     SessionStatus
}

func NewSession() *Session {
	return &Session{
		OutChannel: make(chan differ.PatchBrowser),
		Status:     SessionNew,
	}
}

func (s *Session) QueueMessage(message differ.PatchBrowser) {
	go func() {
		s.OutChannel <- message
	}()
}

func (s *Session) IngestMessage(message BrowserEvent) error {

	defer func() {
		payload := recover()
		if payload != nil {
			// TODO: get session key in log
			s.log(golive.LogWarn, fmt.Sprintf("ingest message: recover from panic: %v", payload), golive.LogEx{"message": message})
		}
	}()

	return s.LivePage.HandleBrowserEvent(message)
}

func (s *Session) ActivatePage(lp *Page) {
	s.LivePage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from page
			evt := <-s.LivePage.Events

			s.log(golive.LogDebug, fmt.Sprintf("component %s triggering %d", evt.Component.Name, evt.Type), golive.LogEx{"evt": evt})

			switch evt.Type {
			case PageComponentUpdated:
				if err := s.LiveRenderComponent(evt.Component, evt.Source); err != nil {
					s.log(golive.LogError, "entryComponent live render", golive.LogEx{"error": err})
				}
				break
			case PageComponentMounted:
				s.QueueMessage(differ.PatchBrowser{
					ComponentID:  evt.Component.Name,
					Type:         EventLiveConnectElement,
					Instructions: nil,
				})
				break
			}
		}
	}()
}

func (s *Session) generateBrowserPatchesFromDiff(diff *differ.Diff, source *EventSource) ([]*differ.PatchBrowser, error) {

	bp := make([]*differ.PatchBrowser, 0)

	for _, instruction := range diff.Instructions {

		selector, err := differ.SelectorFromNode(instruction.Element)
		if skipUpdateValueOnInput(instruction, source) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("selector from node: %w instruction: %v", err, instruction)
		}

		componentID, err := componentIDFromNode(instruction.Element)

		if err != nil {
			return nil, err
		}

		var patch *differ.PatchBrowser

		// find if there is already a patch
		for _, pb := range bp {
			if pb.ComponentID == componentID {
				patch = pb
				break
			}
		}

		// If there is no patch
		if patch == nil {
			patch = differ.NewPatchBrowser(componentID)
			patch.Type = EventLiveDom
			bp = append(bp, patch)
		}

		patch.AddInstruction(differ.PatchInstruction{
			Name: EventLiveDom,
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

func skipUpdateValueOnInput(in differ.ChangeInstruction, source *EventSource) bool {
	if in.Element == nil || source == nil || in.ChangeType != differ.SetAttr || strings.ToLower(in.Attr.Name) != "value" {
		return false
	}

	attr := differ.GetAttribute(in.Element, GoLiveInput)

	return attr != nil && source.Type == EventSourceInput && attr.Val == source.Value
}

// LiveRenderComponent render the updated Component and compare with
// last state. It may apply with *all child components*
func (s *Session) LiveRenderComponent(c *Component, source *EventSource) error {
	var err error

	diff, err := c.LiveRender()

	if err != nil {
		return err
	}

	patches, err := s.generateBrowserPatchesFromDiff(diff, source)

	if err != nil {
		return err
	}

	for _, om := range patches {
		s.QueueMessage(*om)
	}

	return nil
}
