package golive

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// AttrMapFromNode todo
func AttrMapFromNode(node *html.Node) map[string]string {
	m := map[string]string{}
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}

// CreateDOMFromString todo
func CreateDOMFromString(data string) (*html.Node, error) {
	reader := bytes.NewReader([]byte(data))

	parent := &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div}

	fragments, err := html.ParseFragmentWithOptions(reader, parent)

	if err != nil {
		return nil, err
	}

	for _, node := range fragments {
		parent.AppendChild(node)
	}

	return parent, nil
}

// RenderNodeToString todo
func RenderNodeToString(e *html.Node) (string, error) {
	var b bytes.Buffer
	err := html.Render(&b, e)

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// RenderNodesToString todo
func RenderNodesToString(nodes []*html.Node) (string, error) {
	text := ""

	for _, node := range nodes {
		rendered, err := RenderNodeToString(node)

		if err != nil {
			return "", err
		}

		text += rendered
	}

	return text, nil
}

func RenderChildren(parent *html.Node) (string, error) {
	return RenderNodesToString(GetChildrenFromNode(parent))
}

func getClassesSeparated(s string) string {
	return strings.Join(strings.Split(strings.TrimSpace(s), " "), ".")
}

func SelfIndexOfNode(n *html.Node) int {
	ix := 0

	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		ix++
	}

	return ix
}

func GetAllChildrenRecursive(n *html.Node) []*html.Node {
	result := make([]*html.Node, 0)

	if n == nil {
		return result
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, c)

		if c != nil {
			result = append(result, GetAllChildrenRecursive(c)...)
		}
	}

	return result
}

// SelectorFromNode
// TODO: Iterate over parents to find key or go-live-component-id to ensure that is unique
func SelectorFromNode(e *html.Node) (string, error) {

	elementSelector := []string{"*"}
	attrs := AttrMapFromNode(e)

	if e.Type == html.ElementNode {

		if attr, ok := attrs["go-live-uid"]; ok {

			elementSelector = append(elementSelector, "[go-live-uid=\"", attr, "\"]")

			if attr, ok := attrs["key"]; ok {
				elementSelector = append(elementSelector, "[key=\"", attr, "\"]")
			}

			selector := strings.Join(elementSelector, "")
			return selector, nil
		}

		if attr, ok := attrs["go-live-component-id"]; ok {
			elementSelector = append(elementSelector, "[go-live-component-id=", attr, "]")
			selector := strings.Join(elementSelector, "")
			return selector, nil
		}

	}

	return "", fmt.Errorf("could not provide a valid selector")
}

// PathToComponentRoot todo
func PathToComponentRoot(e *html.Node) []int {

	path := make([]int, 0)

	for parent := e; parent != nil; parent = parent.Parent {

		attrs := AttrMapFromNode(parent)

		path = append([]int{SelfIndexOfNode(parent)}, path...)

		if _, ok := attrs["go-live-component-id"]; ok {
			return path
		}
	}

	return path
}
