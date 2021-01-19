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
	d.diffNodes(nodeChildren(actual), nodeChildren(proposed))
}

func (d *Diff) diffNodes(actualNodes, proposedNodes []*html.Node) {

actual:
	for actualNodeIndex, actualNode := range actualNodes {
		for proposedNodeIndex, proposedNode := range proposedNodes {

			// if there is any text change in proposed element nodes
			// the entire parent node will be updated!
			if actualNodeIndex == proposedNodeIndex && (proposedNode.Type == html.TextNode || actualNode.Type == html.TextNode) {
				d.checkpoint()
				d.diffNodeText(actualNode, proposedNode)
				if d.hasChanged() {
					return
				}
				return
			}

			if isSameElementRef(actualNode, proposedNode) {
				continue actual
			}
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
		// TODO: move this to a separated function
		for indexActual, actualNode := range actualNodes {
			if isSameElementRef(actualNode, proposedNode) {
				// If the element is the same but with different index
				// this element should be moved
				if indexActual != proposedIndex {
					d.instructions = append(d.instructions, changeInstruction{
						changeType: Move,
						element:    actualNode,
						index:      indexActual,
					})
					d.diffNode(actualNode, proposedNode)
				}
				continue proposed
			}
		}

		// At this point, means that the proposed element does
		// not exist already. Need to be created
		d.instructions = append(d.instructions, changeInstruction{
			changeType: Append,
			element:    proposedNode,
			index:      proposedIndex,
		})
	}

}

func (d *Diff) diffNodeText(actual, proposed *html.Node) {

	if proposed.Type != html.TextNode || actual.Type != html.TextNode {
		// It is not text
		return
	}

	if proposed.Data == actual.Data {
		// There is no diff
		return
	}

	renderedInnerHTML, _ := renderChildrenNodes(proposed.Parent)

	d.instructions = append(d.instructions, changeInstruction{
		changeType: SetInnerHTML,
		content:    renderedInnerHTML,
		element:    proposed.Parent,
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
	return !(node.Type == html.TextNode && len(strings.TrimSpace(node.Data)) == 0)
}

func isSameElementRef(a, b *html.Node) bool {
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
