package golive

import (
	"golang.org/x/net/html"
	"strconv"
)

type DiffType int

func (dt DiffType) String() string {
	return strconv.Itoa(int(dt))
}

const (
	Append DiffType = iota
	Remove
	SetInnerHtml
	SetAttr
	RemoveAttr
	Replace
)

// ChangeInstruction todo
type ChangeInstruction struct {
	Type    DiffType
	Element *html.Node
	Content string
	Attr    AttrChange
}

// AttrChange todo
type AttrChange struct {
	Name  string
	Value string
}

type Diff struct {
	actual *html.Node

	instructions []ChangeInstruction
	quantity     int
}

func NewDiff(actual *html.Node) *Diff {
	return &Diff{
		actual:       actual,
		instructions: make([]ChangeInstruction, 0),
	}
}

// InstructionsByType todo
func (d *Diff) InstructionsByType(t DiffType) []ChangeInstruction {
	s := make([]ChangeInstruction, 0)

	for _, i := range d.instructions {
		if i.Type == t {
			s = append(s, i)
		}
	}

	return s
}

func (d *Diff) ChangeCheck() {
	d.quantity = len(d.instructions)
}

func (d *Diff) HasChanged() bool {
	return len(d.instructions) != d.quantity
}

// Propose todo
func (d *Diff) Propose(proposed *html.Node) {
	actualChildren := NodeChildren(d.actual)
	proposedChildren := NodeChildren(proposed)
	d.DiffNodes(actualChildren, proposedChildren)
}

func (d *Diff) DiffNode(actual, proposed *html.Node) {

	if actual.Data != proposed.Data {
		content, _ := RenderNodeToString(proposed)
		d.instructions = append(d.instructions, ChangeInstruction{
			Type:    Replace,
			Element: actual,
			Content: content,
		})
		return
	}

	d.DiffNodeAttributes(actual, proposed)

	/**
	If is a text node and has some difference between them
	so, we'll be replacing the entire content of parent
	- So, we recommend you proposed always set the reactive
	  text proposed be inside of any dom element
	*/
	if proposed.Type == html.TextNode {
		d.ChangeCheck()
		d.DiffNodeText(actual, proposed)
		if d.HasChanged() {
			return
		}
	}

	if !IsChildrenTheSame(proposed, actual) {
		d.DiffNodes(NodeChildren(actual), NodeChildren(proposed))
	}
}

// DiffNodes todo
func (d *Diff) DiffNodes(actual, proposed []*html.Node) {
	actualLen := len(actual)
	proposedLen := len(proposed)

	// TODO: comment why
	minLen := actualLen

	// Iterate over all the proposed nodes
	// And verify is have some text change
	d.ChangeCheck()
	for index, proposedNode := range proposed {
		if proposedNode.Type == html.TextNode {

			actualNode := &html.Node{}

			// node index exists in actual?
			if index < len(actual) {
				actualNode = actual[index]
			}

			d.DiffNodeText(actualNode, proposedNode)
		}
	}

	// If some text has been changed
	// the entire innerHTML will be replaced
	if d.HasChanged() {
		return
	}

	if actualLen < proposedLen {
		// Get all spare nodes
		proposedAppendNodes := proposed[actualLen:]

		for _, proposedToAppendNode := range proposedAppendNodes {
			renderedOuterHTML, _ := RenderNodeToString(proposedToAppendNode)
			d.instructions = append(d.instructions, ChangeInstruction{
				Type:    Append,
				Element: proposedToAppendNode.Parent,
				Content: renderedOuterHTML,
			})
		}

		minLen = actualLen
	}

	if actualLen > proposedLen {

		// Remove the resting nodes
		toRemoveNodes := actual[proposedLen:]

		for _, node := range toRemoveNodes {

			if node.Type == html.TextNode {
				// empty text node is needed because
				// it will generate a diff instruction
				d.DiffNodeText(node, &html.Node{Type: html.TextNode})
				break
			}

			d.instructions = append(d.instructions, ChangeInstruction{
				Type:    Remove,
				Element: node,
			})
		}
		minLen = proposedLen
	}

	// Diff children
	for i := 0; i < minLen; i++ {

		actualNode := actual[i]
		proposedNode := proposed[i]

		d.DiffNode(actualNode, proposedNode)
	}
}

// DiffNodeText todo
func (d *Diff) DiffNodeText(actual, proposed *html.Node) {

	if proposed.Type != html.TextNode {
		// It is not text
		return
	}

	if proposed.Data == actual.Data {
		// There is no diff
		return
	}

	renderedInnerHTML, _ := RenderChildrenNodes(proposed.Parent)

	node := actual

	if node.Parent == nil {
		node = proposed
	}

	d.instructions = append(d.instructions, ChangeInstruction{
		Type:    SetInnerHtml,
		Content: renderedInnerHTML,
		Element: node.Parent,
	})
}

// DiffNodeAttributes compares the attributes in el to the attributes in otherEl
// and adds the necessary patches to make the attributes in el match those in
// otherEl
func (d *Diff) DiffNodeAttributes(from, to *html.Node) {

	otherAttrs := AttrMapFromNode(to)
	attrs := AttrMapFromNode(from)

	// Now iterate through the attributes in otherEl
	for name, otherValue := range otherAttrs {
		value, found := attrs[name]
		if !found || value != otherValue {

			d.instructions = append(d.instructions, ChangeInstruction{
				Type:    SetAttr,
				Element: from,
				Attr: AttrChange{
					Name:  name,
					Value: otherValue,
				},
			})
		}
	}

	for attrName := range attrs {
		if _, found := otherAttrs[attrName]; !found {

			d.instructions = append(d.instructions, ChangeInstruction{
				Type:    RemoveAttr,
				Element: from,
				Attr: AttrChange{
					Name: attrName,
				},
			})
		}
	}

}
