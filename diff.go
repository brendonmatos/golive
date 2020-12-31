package golive

import (
	"fmt"
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
	Type        DiffType
	Element     *html.Node
	Content     string
	Attr        interface{}
	ComponentId string
}

// Attr todo
type Attr struct {
	Name  string
	Value string
}

// GetDiffFromNodes todo
func GetDiffFromNodes(start, end *html.Node) []ChangeInstruction {
	instructions := make([]ChangeInstruction, 0)
	RecursiveDiff(&instructions, GetChildrenFromNode(start), GetChildrenFromNode(end))
	return instructions
}

func componentIdFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {

		attrs := AttrMapFromNode(parent)

		if parent.Type == html.ElementNode {
			if id, ok := attrs["go-live-component-id"]; ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("node not found")
}

// GetChildrenFromNode todo
func GetChildrenFromNode(n *html.Node) []*html.Node {
	children := make([]*html.Node, 0)

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}

	return children
}

// RecursiveDiff todo
func RecursiveDiff(changeList *[]ChangeInstruction, from, to []*html.Node) {

	fromLen := len(from)
	toLen := len(to)
	minLen := fromLen

	if fromLen < toLen {
		// take the new nodes
		toAppendNodes := to[fromLen:]

		for _, node := range toAppendNodes {
			rendered, _ := RenderNodeToString(node)
			componentId, _ := componentIdFromNode(node)
			*changeList = append(*changeList, createChangeInstruction(Append, componentId, node.Parent, rendered, nil))
		}

		minLen = fromLen
	}

	if fromLen > toLen {
		// take the excess
		toRemoveNodes := from[toLen:]

		for _, node := range toRemoveNodes {
			componentId, _ := componentIdFromNode(node)
			*changeList = append(*changeList, createChangeInstruction(Remove, componentId, node, "", nil))
		}
		minLen = toLen
	}

	// Diff children
	for i := 0; i < minLen; i++ {
		fromNode := from[i]
		toNode := to[i]

		AttributesDiff(changeList, fromNode, toNode)

		prevLen := len(*changeList)

		TextDiff(changeList, fromNode, toNode)

		if len(*changeList) > prevLen {
			continue
		}

		/**
		If is a text node and has some difference between them
		so, we'll be replacing the entire content of parent
		- So, we recommend you to always set the reactive
		  text to be inside of any dom element
		*/
		if !IsChildrenTheSame(toNode, fromNode) {
			RecursiveDiff(changeList, GetChildrenFromNode(fromNode), GetChildrenFromNode(toNode))
		}

	}

}

func createChangeInstruction(diffType DiffType, componentId string, el *html.Node, rendered string, attr *Attr) ChangeInstruction {
	return ChangeInstruction{
		Type:        diffType,
		Element:     el,
		ComponentId: componentId,
		Content:     rendered,
		Attr:        attr,
	}
}

func TextDiff(changeList *[]ChangeInstruction, from, to *html.Node) {

	if to.Type != html.TextNode {
		// It is not text
		return
	}

	if to.Data == from.Data {
		// There is no diff
		return
	}

	parent := to.Parent
	componentId, _ := componentIdFromNode(parent)
	rendered := RenderChildren(parent)
	*changeList = append(*changeList, createChangeInstruction(SetInnerHtml, componentId, to, rendered, nil))
}

// AttributesDiff compares the attributes in el to the attributes in otherEl
// and adds the necessary patches to make the attributes in el match those in
// otherEl
func AttributesDiff(changeList *[]ChangeInstruction, from, to *html.Node) {
	otherAttrs := AttrMapFromNode(to)
	attrs := AttrMapFromNode(from)

	// Now iterate through the attributes in otherEl
	for name, otherValue := range otherAttrs {
		value, found := attrs[name]
		if !found || value != otherValue {

			componentId, _ := componentIdFromNode(from)

			*changeList = append(*changeList, createChangeInstruction(SetAttr, componentId, from, "", &Attr{
				Name:  name,
				Value: otherValue,
			}))
		}
	}

	for attrName := range attrs {
		if _, found := otherAttrs[attrName]; !found {
			componentId, _ := componentIdFromNode(from)
			*changeList = append(*changeList, createChangeInstruction(RemoveAttr, componentId, from, "", &Attr{
				Name: attrName,
			}))
		}
	}

}

// IsChildrenTheSame todo
func IsChildrenTheSame(n *html.Node, other *html.Node) bool {
	return RenderNodesToString(GetChildrenFromNode(n)) == RenderNodesToString(GetChildrenFromNode(other))
}

// GetDiffFromRawHTML todo
func GetDiffFromRawHTML(start string, end string) ([]ChangeInstruction, error) {
	startN, err := CreateDOMFromString(start)

	if err != nil {
		return nil, err
	}

	endN, err := CreateDOMFromString(end)

	if err != nil {
		return nil, err
	}

	return GetDiffFromNodes(startN, endN), nil
}
