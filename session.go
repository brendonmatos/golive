package golive

import (
	"strconv"
)

const (
	EventLiveInput          = "li"
	EventLiveMethod         = "lm"
	EventLiveDom            = "ld"
	EventLiveConnectElement = "lce"
	EventLiveDisconnect     = "lx"
)

type BrowserEvent struct {
	Name         string `json:"name"`
	ComponentID  string `json:"component_id"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
	StateKey     string `json:"key"`
	StateValue   string `json:"value"`
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
					s.log(LogError, "component live render", logEx{"error": err})
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

		selector, err := SelectorFromNode(instruction.Element)

		if err != nil {
			return nil, err
		}

		componentId, err := ComponentIdFromNode(instruction.Element)

		if err != nil {
			return nil, err
		}

		var patch *PatchBrowser

		// find if there is already a patch
		for _, pb := range bp {
			if pb.ComponentID == componentId {
				patch = pb
				break
			}
		}

		// IF there is no patch
		if patch == nil {
			patch = NewPatchBrowser(componentId)
			patch.Name = EventLiveDom
			bp = append(bp, patch)
		}

		patch.AddInstruction(PatchInstruction{
			Name:     EventLiveDom,
			Type:     strconv.Itoa(int(instruction.Type)),
			Attr:     instruction.Attr,
			Content:  instruction.Content,
			Selector: selector,
		})
	}
	return bp, nil
}

// LiveRenderComponent render the updated component and compare with
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
