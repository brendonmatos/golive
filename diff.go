package golive

import (
	"golang.org/x/net/html"
	"regexp"
)

var spaceRegex = regexp.MustCompile(`\s+`)

type DiffType int

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
	actual       *html.Node
	instructions []ChangeInstruction
}

func NewDiff(actual *html.Node) *Diff {
	return &Diff{
		actual:       actual,
		instructions: make([]ChangeInstruction, 0),
	}
}

func (d *Diff) Propose(proposed *html.Node) {

	actualChildren := GetChildrenFromNode(d.actual)
	proposedChildren := GetChildrenFromNode(proposed)

	d.DiffBetweenNodes(actualChildren, proposedChildren)
}

func (d *Diff) DiffBetweenNodes(actual, proposed []*html.Node) {
	actualLen := len(actual)
	proposedLen := len(proposed)

	minLen := actualLen

	// Iterate over all the proposed nodes
	// And verify is have some text change
	clSize := len(d.instructions)
	for index, proposedNode := range proposed {
		if proposedNode.Type == html.TextNode {

			actualNode := &html.Node{}

			// node index exists in actual?
			if index <= actualLen {
				actualNode = actual[index]
			}

			d.DiffBetweenText(actualNode, proposedNode)
		}
	}

	// If some text has been changed
	// the entire node will be replaced
	if len(d.instructions) > clSize {
		return
	}

	if actualLen < proposedLen {
		toAppendNodes := proposed[actualLen:]

		for _, node := range toAppendNodes {
			rendered, _ := RenderNodeToString(node)
			d.instructions = append(d.instructions, ChangeInstruction{
				Type:    Append,
				Element: node.Parent,
				Content: rendered,
			})
		}

		minLen = actualLen
	}

	if actualLen > proposedLen {

		// Remove the resting nodes
		toRemoveNodes := actual[proposedLen:]

		for _, node := range toRemoveNodes {

			if node.Type == html.TextNode {
				d.DiffBetweenText(&html.Node{}, node)
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

		fromNode := actual[i]
		toNode := proposed[i]

		d.DiffNodeAttributes(fromNode, toNode)

		/**
		If is a text node and has some difference between them
		so, we'll be replacing the entire content of parent
		- So, we recommend you proposed always set the reactive
		  text proposed be inside of any dom element
		*/
		if toNode.Type == html.TextNode {
			d.DiffBetweenText(fromNode, toNode)
		} else if !IsChildrenTheSame(toNode, fromNode) {
			if fromNode.Type == html.TextNode {
				continue
			}
			if toNode.Type == html.TextNode {
				continue
			}
			d.DiffBetweenNodes(GetChildrenFromNode(fromNode), GetChildrenFromNode(toNode))
		}
	}
}

func (d *Diff) DiffBetweenText(from, to *html.Node) {

	if to.Type != html.TextNode {
		// It is not text
		return
	}

	if to.Data == from.Data {
		// There is no diff
		return
	}

	parent := to.Parent
	rendered, _ := RenderNodeChildren(parent)

	d.instructions = append(d.instructions, ChangeInstruction{
		Type:    SetInnerHtml,
		Content: rendered,
		Element: parent,
	})
}

func (d *Diff) DiffNodeAttributes(from, to *html.Node) {
	// AttributesDiff compares the attributes in el to the attributes in otherEl
	// and adds the necessary patches to make the attributes in el match those in
	// otherEl
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

// IsChildrenTheSame todo

//// GetDiffFromRawHTML todo
//func GetDiffFromRawHTML(start string, end string) ([]ChangeInstruction, error) {
//	startN, err := CreateDOMFromString(start)
//
//	if err != nil {
//		return nil, err
//	}
//
//	endN, err := CreateDOMFromString(end)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return GetDiffFromNodes(startN, endN), nil
//}
