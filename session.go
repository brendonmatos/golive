package golive

import (
	"fmt"
	"strconv"
)

const (
	EventLiveInput          = "li"
	EventLiveMethod         = "lm"
	EventLiveDom            = "ld"
	EventLiveDisconnect     = "lx"
	EventLiveError          = "le"
	EventLiveConnectElement = "lce"
)

var (
	LiveErrorSessionNotFound = "session_not_found"
)

func LiveErrorMap() map[string]string {
	return map[string]string{
		"LiveErrorSessionNotFound": LiveErrorSessionNotFound,
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

type OutMessage struct {
	Name        string      `json:"name"`
	ComponentID string      `json:"component_id"`
	Type        string      `json:"type"`
	Attr        interface{} `json:"attr,omitempty"`
	Content     string      `json:"content,omitempty"`
	Element     string      `json:"element"`
}

type Session struct {
	LivePage   *Page
	OutChannel chan PatchBrowser
	log        Log
}

func NewSession() *Session {
	return &Session{
		OutChannel: make(chan PatchBrowser),
	}
}

func (s *Session) QueueMessage(message PatchBrowser) {
	go func() {
		s.OutChannel <- message
	}()
}

func (s *Session) IngestMessage(message BrowserEvent) error {

	defer func() {
		payload := recover()
		if payload != nil {
			// TODO: get session key in log
			s.log(LogWarn, fmt.Sprintf("ingest message: recover from panic: %v", payload), logEx{"message": message})
		}
	}()

	err := s.LivePage.HandleBrowserEvent(message)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) ActivatePage(lp *Page) {
	s.LivePage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from page
			evt := <-s.LivePage.Events

			switch evt.Type {
			case PageComponentUpdated:
				if err := s.LiveRenderComponent(evt.Component); err != nil {
					s.log(LogError, "entryComponent live render", logEx{"error": err})
				}
				break
			case PageComponentMounted:
				s.QueueMessage(PatchBrowser{
					ComponentID:  evt.Component.Name,
					Name:         EventLiveConnectElement,
					Instructions: nil,
				})
				break
			}
		}
	}()
}

func (s *Session) generateBrowserPatchesFromDiff(diff *Diff) ([]*PatchBrowser, error) {

	bp := make([]*PatchBrowser, 0)

	for _, instruction := range diff.instructions {

		selector, err := selectorFromNode(instruction.element)

		if err != nil {
			return nil, fmt.Errorf("selector from node: %w instruction: %v", err, instruction)
		}

		componentID, err := ComponentIDFromNode(instruction.element)

		if err != nil {
			return nil, err
		}

		var patch *PatchBrowser

		// find if there is already a patch
		for _, pb := range bp {
			if pb.ComponentID == componentID {
				patch = pb
				break
			}
		}

		// IF there is no patch
		if patch == nil {
			patch = NewPatchBrowser(componentID)
			patch.Name = EventLiveDom
			bp = append(bp, patch)
		}

		patch.AddInstruction(PatchInstruction{
			Name:     EventLiveDom,
			Type:     strconv.Itoa(int(instruction.changeType)),
			Attr:     instruction.attr,
			Content:  instruction.content,
			Selector: selector.toString(),
		})
	}
	return bp, nil
}

// LiveRenderComponent render the updated entryComponent and compare with
// last state. It may apply with *all child components*
func (s *Session) LiveRenderComponent(c *LiveComponent) error {
	var err error

	diff, err := c.LiveRender()

	if err != nil {
		return err
	}

	patches, err := s.generateBrowserPatchesFromDiff(diff)

	if err != nil {
		return err
	}

	for _, om := range patches {
		s.QueueMessage(*om)
	}

	return nil
}
