package golive

import (
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

type DiffType int

func (dt DiffType) String() string {
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

type Diff struct {
	actual       *html.Node
	instructions []changeInstruction
	quantity     int
}

func NewDiff(actual *html.Node) *Diff {
	return &Diff{
		actual:       actual,
		instructions: make([]changeInstruction, 0),
	}
}

func (d *Diff) instructionsByType(t DiffType) []changeInstruction {
	s := make([]changeInstruction, 0)

	for _, i := range d.instructions {
		if i.changeType == t {
			s = append(s, i)
		}
	}

	return s
}

func (d *Diff) checkpoint() {
	d.quantity = len(d.instructions)
}

// Has changed since last checkpoint
func (d *Diff) hasChanged() bool {
	return len(d.instructions) != d.quantity
}

func (d *Diff) propose(proposed *html.Node) {
	d.diffNode(d.actual, proposed)
}

func (d *Diff) diffNode(actual, proposed *html.Node) {

	if actual.Data != proposed.Data {
		content, _ := renderNodeToString(proposed)
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Replace,
			element:    actual,
			content:    content,
		})
		return
	}

	d.diffNodeAttributes(actual, proposed)
	d.diffChildren(actual, proposed)
}

func (d *Diff) diffChildren(actualParent, proposedParent *html.Node) {
	actualNodes := nodeChildren(actualParent)
	proposedNodes := nodeChildren(proposedParent)

actual:
	for actualIndex, actualNode := range actualNodes {
		for proposedIndex, proposedNode := range proposedNodes {
			// TODO: comment why i'm skipping this here
			if actualIndex == proposedIndex || hasSameElementRef(actualNode, proposedNode) {
				continue actual
			}

			if !nodeRelevant(proposedNode) {
				continue actual
			}
		}

		// If reach here, mean that the text node does
		// not exists in proposedParent so, render the entire parent content
		// removing
		if actualNode.Type == html.TextNode {
			d.forceRenderElementContent(proposedParent)
			return
		}

		d.instructions = append(d.instructions, changeInstruction{
			changeType: Remove,
			element:    actualNode,
		})
	}

proposed:
	for proposedIndex, proposedNode := range proposedNodes {

		// This part will be used in case of the element changed
		// index. The actualIndex and proposedIndex should never
		// be equal at this moment
		for actualIndex, actualNode := range actualNodes {

			if actualIndex == proposedIndex {
				if actualNode.Type == html.TextNode || proposedNode.Type == html.TextNode {
					// place a checkpoint
					d.checkpoint()

					if !nodeRelevant(actualNode) && !nodeRelevant(proposedNode) {
						continue proposed
					}

					// differentiate two text nodes
					d.diffNodeText(actualNode, proposedNode)

					// has something changed?
					if d.hasChanged() {
						return
					} else {
						continue proposed
					}
				} else if proposedNode.Type == html.ElementNode {
					d.diffNode(actualNode, proposedNode)
					continue proposed
				}
			} else if hasSameElementRef(actualNode, proposedNode) {

				// If the element is the same but with different index
				// this element should be moved
				if actualIndex != proposedIndex {
					d.instructions = append(d.instructions, changeInstruction{
						changeType: Move,
						element:    actualNode,
						index:      actualIndex,
					})
				}
				continue proposed
			}

		}

		if proposedNode.Type == html.TextNode {
			d.forceRenderElementContent(proposedParent)
			return
		}

		// At this point, means that the proposedParent element does
		// not exist already. Need to be created
		nodeContent, _ := renderNodeToString(proposedNode)
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Append,
			element:    proposedNode.Parent,
			content:    nodeContent,
			index:      proposedIndex,
		})
	}

}

func (d *Diff) diffNodeText(actual, proposed *html.Node) {

	if actual == nil || proposed == nil || actual.Data == proposed.Data {
		return
	}

	d.forceRenderElementContent(proposed.Parent)
}

func (d *Diff) forceRenderElementContent(proposed *html.Node) {
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
func (d *Diff) diffNodeAttributes(actual, proposed *html.Node) {

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

func nodeRelevant(node *html.Node) bool {
	if node == nil {
		return false
	}

	if node.Type == html.TextNode && len(strings.TrimSpace(node.Data)) == 0 {
		return false
	}

	return true
}

func hasSameElementRef(a, b *html.Node) bool {
	var err error

	aSelector, err := selectorFromNode(a)

	if err != nil || aSelector == nil {
		return false
	}

	bSelector, err := selectorFromNode(b)

	if err != nil || bSelector == nil {
		return false
	}

	return aSelector.toString() == bSelector.toString()
}
