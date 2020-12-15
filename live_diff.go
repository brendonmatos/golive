package golive

import (
	"fmt"
	"golang.org/x/net/html"
	"regexp"
)

var spaceRegex = regexp.MustCompile(`\s+`)

// ChangeInstruction todo
type ChangeInstruction struct {
	Type    string
	Element string
	Content string
	Attr    interface{}
	ScopeID string
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

func ScopeOfNode(e *html.Node) (string, error) {
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
		toAppendNodes := to[fromLen:]

		for _, node := range toAppendNodes {
			scopeID, _ := ScopeOfNode(node)
			*changeList = append(*changeList, ChangeInstruction{
				Type:    "APPEND",
				Element: SelectorFromNode(node.Parent),
				Content: RenderNodeToString(node),
				ScopeID: scopeID,
			})
		}

		minLen = fromLen
	}

	if fromLen > toLen {
		toRemoveNodes := from[toLen:]

		for _, node := range toRemoveNodes {
			scopeID, _ := ScopeOfNode(node)

			*changeList = append(*changeList, ChangeInstruction{
				Type:    "REMOVE",
				Element: SelectorFromNode(node),
				ScopeID: scopeID,
			})
		}
		minLen = toLen
	}

	for i := 0; i < minLen; i++ {

		fromNode := from[i]
		toNode := to[i]

		if toNode.Type == html.TextNode {

			if toNode.Data != fromNode.Data {

				scopeID, _ := ScopeOfNode(fromNode)
				*changeList = append(*changeList, ChangeInstruction{
					Type:    "SET_INNER_HTML",
					Content: RenderNodeToString(toNode),
					Element: SelectorFromNode(fromNode),
					ScopeID: scopeID,
				})
			}

		} else if !IsChildrenTheSame(toNode, fromNode) {
			RecursiveDiff(changeList, GetChildrenFromNode(fromNode), GetChildrenFromNode(toNode))
		}

		AttributesDiff(changeList, fromNode, toNode)
	}

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

			scopeID, _ := ScopeOfNode(from)
			*changeList = append(*changeList, ChangeInstruction{
				Type:    "SET_ATTR",
				ScopeID: scopeID,
				Element: SelectorFromNode(from),
				Attr: Attr{
					Name:  name,
					Value: otherValue,
				},
			})
		}
	}

	for attrName := range attrs {
		if _, found := otherAttrs[attrName]; !found {

			scopeID, _ := ScopeOfNode(from)
			*changeList = append(*changeList, ChangeInstruction{
				Type:    "REMOVE_ATTR",
				ScopeID: scopeID,
				Element: SelectorFromNode(from),
				Attr: Attr{
					Name: attrName,
				},
			})
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
