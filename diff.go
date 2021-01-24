package golive

import (
	"golang.org/x/net/html"
	"strconv"
	"strings"
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

	d.markNodeDone(proposed)

	uidActual, actualOk := getLiveUidAttributeValue(actual)
	uidProposed, proposedOk := getLiveUidAttributeValue(proposed)

	if actualOk && proposedOk {
		if uidActual != uidProposed {
			content, _ := renderNodeToString(proposed)
			d.instructions = append(d.instructions, changeInstruction{
				changeType: Replace,
				element:    actual,
				content:    content,
			})
			return
		}
	}

	d.diffNodeAttributes(actual, proposed)
	d.diffChildren(actual, proposed)
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

func (d *diff) diffChildren(actualParent, proposedParent *html.Node) {

	proposedChildIndex := -1
	for proposedChild := proposedParent.FirstChild; proposedChild != nil; proposedChild = proposedChild.NextSibling {

		proposedChildIndex++

		relativeActualChild := getChildNodeIndex(actualParent, proposedChildIndex)

		if relativeActualChild == nil {
			if proposedChild.Type == html.TextNode {
				d.forceRenderElementContent(proposedParent)
				d.markNodeDone(proposedParent)
				return
			}

			if proposedChild.Type == html.ElementNode {
				// At this point, means that the proposedParent element does
				// not exist already. Need to be created
				nodeContent, _ := renderNodeToString(proposedChild)
				d.instructions = append(d.instructions, changeInstruction{
					changeType: Append,
					element:    proposedParent,
					content:    nodeContent,
				})
				d.markNodeDone(proposedParent)
			}
		} else {

			d.diffNode(relativeActualChild, proposedChild)
		}
	}

	actualChildIndex := -1
	for actualChild := actualParent.FirstChild; actualChild != nil; actualChild = actualChild.NextSibling {

		actualChildIndex++

		relativeProposedChild := getChildNodeIndex(actualParent, actualChildIndex)

		if !nodeRelevant(actualChild) {
			continue
		}

		if relativeProposedChild == nil {
			if actualChild.Type == html.ElementNode {
				d.instructions = append(d.instructions, changeInstruction{
					changeType: Remove,
					element:    actualChild,
				})
				d.markNodeDone(actualChild)
			}

			if actualChild.Type == html.TextNode {
				d.forceRenderElementContent(proposedParent)
				d.markNodeDone(proposedParent)
				return
			}
		}
	}

	actualNodes := nodeChildrenElements(actualParent)
	proposedNodes := nodeChildrenElements(proposedParent)

	// The main purpose to this loop is basically to
	// Remove sobresalent nodes and move elmeents
actual:
	for actualIndex, actualNode := range actualNodes {
		for proposedIndex, proposedNode := range proposedNodes {

			if d.isMarked(proposedNode) {
				continue actual
			}

			hasSameRef := hasSameElementRef(actualNode, proposedNode)

			if hasSameRef {

				d.diffNode(actualNode, proposedNode)

				if actualIndex != proposedIndex {
					// Means that the element is changed of position
					d.instructions = append(d.instructions, changeInstruction{
						changeType: Move,
						element:    proposedNode,
						index:      proposedIndex,
					})
				}

				continue actual
			}
		}

		d.instructions = append(d.instructions, changeInstruction{
			changeType: Remove,
			element:    actualNode,
		})
		d.markNodeDone(actualNode)
	}

}

func (d *diff) diffNodeText(actual, proposed *html.Node) {

	if actual == nil || proposed == nil || actual.Data == proposed.Data {
		return
	}

	d.forceRenderElementContent(proposed.Parent)
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

func nodeRelevant(node *html.Node) bool {
	if node == nil {
		return false
	}

	if node.Type == html.TextNode && len(strings.TrimSpace(node.Data)) == 0 {
		return false
	}

	return true
}

func getChildNodeIndex(node *html.Node, index int) *html.Node {

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if index == 0 {
			return child
		}
		index--
	}
	return nil
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
