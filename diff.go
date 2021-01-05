package golive

import (
	"fmt"
	"regexp"

	"golang.org/x/net/html"
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
	componentId string
}

// Attr todo
type Attr struct {
	Name  string
	Value string
}

// GetDiffFromNodes todo
func GetDiffFromNodes(start, end *html.Node) []ChangeInstruction {
	instructions := make([]ChangeInstruction, 0)
	childrenFrom := GetChildrenFromNode(start)
	childrenTo := GetChildrenFromNode(end)
	RecursiveDiff(&instructions, childrenFrom, childrenTo)
	return instructions
}

func ComponentIdFromNode(e *html.Node) (string, error) {
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
func RecursiveDiff(changeList *[]ChangeInstruction, actual, proposed []*html.Node) {
	actualLen := len(actual)
	proposedLen := len(proposed)
	minLen := actualLen

	// Iterate over all the proposed nodes
	// And verify is have some text change
	clSize := len(*changeList)

	for index, proposedNode := range proposed {
		if proposedNode.Type == html.TextNode {
			fromNode := &html.Node{}

			if index < len(actual) {
				fromNode = actual[index]
			}

			TextDiff(changeList, fromNode, proposedNode)
		}
	}

	// If there is a change, return. Because the entire
	// Node will be replaced
	if len(*changeList) > clSize {
		return
	}

	if actualLen < proposedLen {
		toAppendNodes := proposed[actualLen:]

		for _, node := range toAppendNodes {
			rendered, _ := RenderNodeToString(node)
			componentId, _ := ComponentIdFromNode(node)
			*changeList = append(*changeList, ChangeInstruction{
				Type:        Append,
				Element:     node.Parent,
				Content:     rendered,
				componentId: componentId,
			})
		}

		minLen = actualLen
	}

	if actualLen > proposedLen {

		toRemoveNodes := actual[proposedLen:]

		for _, node := range toRemoveNodes {

			if node.Type == html.TextNode {
				TextDiff(changeList, &html.Node{}, node)
				break
			}

			componentId, _ := ComponentIdFromNode(node)

			*changeList = append(*changeList, ChangeInstruction{
				Type:        Remove,
				Element:     node,
				componentId: componentId,
			})
		}
		minLen = proposedLen
	}

	// Diff children
	for i := 0; i < minLen; i++ {

		fromNode := actual[i]
		toNode := proposed[i]

		AttributesDiff(changeList, fromNode, toNode)

		/**
		If is a text node and has some difference between them
		so, we'll be replacing the entire content of parent
		- So, we recommend you to always set the reactive
		  text to be inside of any dom element
		*/
		if toNode.Type == html.TextNode {
			TextDiff(changeList, fromNode, toNode)
		} else if !IsChildrenTheSame(toNode, fromNode) {
			if fromNode.Type == html.TextNode {
				continue
			}
			if toNode.Type == html.TextNode {
				continue
			}
			RecursiveDiff(changeList, GetChildrenFromNode(fromNode), GetChildrenFromNode(toNode))
		}

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
	componentId, _ := ComponentIdFromNode(parent)
	rendered, _ := RenderChildren(parent)

	*changeList = append(*changeList, ChangeInstruction{
		Type:        SetInnerHtml,
		Content:     rendered,
		Element:     parent,
		componentId: componentId,
	})
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

			componentId, _ := ComponentIdFromNode(from)
			*changeList = append(*changeList, ChangeInstruction{
				Type:        SetAttr,
				componentId: componentId,
				Element:     from,
				Attr: Attr{
					Name:  name,
					Value: otherValue,
				},
			})
		}
	}

	for attrName := range attrs {
		if _, found := otherAttrs[attrName]; !found {

			componentId, _ := ComponentIdFromNode(from)
			*changeList = append(*changeList, ChangeInstruction{
				Type:        RemoveAttr,
				componentId: componentId,
				Element:     from,
				Attr: Attr{
					Name: attrName,
				},
			})
		}
	}

}

// IsChildrenTheSame todo

func IsChildrenTheSame(n *html.Node, other *html.Node) bool {
	actual, _ := RenderNodesToString(GetChildrenFromNode(n))
	proposed, _ := RenderNodesToString(GetChildrenFromNode(other))
	return actual == proposed
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
