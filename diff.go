package golive

import (
	"golang.org/x/net/html"
	"strconv"
)

type DiffType int

func (dt DiffType) toString() string {
	return strconv.Itoa(int(dt))
}

const (
	Append DiffType = iota
	Remove
	SetInnerHTML
	SetAttr
	RemoveAttr
	Replace
	Move
)

type changeInstruction struct {
	changeType DiffType
	element    *html.Node
	content    string
	attr       attrChange
	index      int
}

// attrChange todo
type attrChange struct {
	name  string
	value string
}

type diff struct {
	actual       *html.Node
	instructions []changeInstruction
	quantity     int
	doneElements []*html.Node
}

func newDiff(actual *html.Node) *diff {
	return &diff{
		actual:       actual,
		instructions: make([]changeInstruction, 0),
	}
}

func (d *diff) instructionsByType(t DiffType) []changeInstruction {
	s := make([]changeInstruction, 0)

	for _, i := range d.instructions {
		if i.changeType == t {
			s = append(s, i)
		}
	}

	return s
}

func (d *diff) checkpoint() {
	d.quantity = len(d.instructions)
}

// Has changed since last checkpoint
func (d *diff) hasChanged() bool {
	return len(d.instructions) != d.quantity
}

func (d *diff) propose(proposed *html.Node) {
	d.clearMarked()
	d.diffNode(d.actual, proposed)
}

func (d *diff) diffNode(actual, proposed *html.Node) {

	uidActual, actualOk := getLiveUidAttributeValue(actual)
	uidProposed, proposedOk := getLiveUidAttributeValue(proposed)

	if actualOk && proposedOk && uidActual != uidProposed {
		content, _ := renderNodeToString(proposed)
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Replace,
			element:    actual,
			content:    content,
		})
		return
	}

	d.diffNodeAttributes(actual, proposed)
	d.diffWalk(actual.FirstChild, proposed.FirstChild)
	d.markNodeDone(proposed)
}

func (d *diff) clearMarked() {
	d.doneElements = make([]*html.Node, 0)
}

func (d *diff) markNodeDone(node *html.Node) {
	d.doneElements = append(d.doneElements, node)
}

func (d *diff) isMarked(node *html.Node) bool {
	for _, n := range d.doneElements {
		if n == node {
			return true
		}
	}

	return false
}

func (d *diff) diffWalk(actual, proposed *html.Node) {

	if actual == nil && proposed == nil {
		return
	}

	if nodeIsText(actual) || nodeIsText(proposed) {
		d.checkpoint()
		d.diffTextNode(actual, proposed)
		if d.hasChanged() {
			return
		}
	}

	if actual != nil && proposed != nil {
		d.diffNode(actual, proposed)
	} else if actual == nil && nodeIsElement(proposed) {
		nodeContent, _ := renderNodeToString(proposed)
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Append,
			element:    proposed.Parent,
			content:    nodeContent,
		})
		d.markNodeDone(proposed)
	} else if proposed == nil && nodeIsElement(actual) {
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Remove,
			element:    actual,
		})
		d.markNodeDone(actual)
	}

	nextActual := nextRelevantElement(actual)
	nextProposed := nextRelevantElement(proposed)

	if nextActual != nil || nextProposed != nil {
		d.diffWalk(nextActual, nextProposed)
	}
}

func (d *diff) forceRenderElementContent(proposed *html.Node) {
	childrenHTML, _ := renderInnerHTML(proposed)

	d.instructions = append(d.instructions, changeInstruction{
		changeType: SetInnerHTML,
		content:    childrenHTML,
		element:    proposed,
	})
}

// diffNodeAttributes compares the attributes in el to the attributes in otherEl
// and adds the necessary patches to make the attributes in el match those in
// otherEl
func (d *diff) diffNodeAttributes(actual, proposed *html.Node) {

	actualAttrs := AttrMapFromNode(actual)
	proposedAttrs := AttrMapFromNode(proposed)

	// Now iterate through the attributes in otherEl
	for name, otherValue := range proposedAttrs {
		value, found := actualAttrs[name]
		if !found || value != otherValue {
			d.instructions = append(d.instructions, changeInstruction{
				changeType: SetAttr,
				element:    actual,
				attr: attrChange{
					name:  name,
					value: otherValue,
				},
			})
		}
	}

	for attrName := range actualAttrs {
		if _, found := proposedAttrs[attrName]; !found {

			d.instructions = append(d.instructions, changeInstruction{
				changeType: RemoveAttr,
				element:    actual,
				attr: attrChange{
					name: attrName,
				},
			})
		}
	}
}

func (d *diff) diffTextNode(actual, proposed *html.Node) {

	// Any node is text
	if !nodeIsText(proposed) && !nodeIsText(actual) {
		return
	}

	proposedIsRelevant := nodeRelevant(proposed)
	actualIsRelevant := nodeRelevant(actual)

	if !proposedIsRelevant && !actualIsRelevant {
		return
	}

	// XOR
	if proposedIsRelevant != actualIsRelevant {
		goto renderEntireNode
	}

	if proposed.Data != actual.Data {
		goto renderEntireNode
	}

	return

renderEntireNode:
	{

		node := proposed

		if node == nil {
			node = actual
		}

		if node == nil {
			return
		}

		d.forceRenderElementContent(node.Parent)
		d.markNodeDone(node.Parent)
	}

}
