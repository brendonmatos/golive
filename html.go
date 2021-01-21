package golive

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	ErrCouldNotProvideValidSelector = fmt.Errorf("could not provide a valid selector")
	ErrElementNotSigned             = fmt.Errorf("element is not signed with go-live-uid")
)

// AttrMapFromNode todo
func AttrMapFromNode(node *html.Node) map[string]string {
	m := map[string]string{}
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}

// nodeFromString todo
func nodeFromString(data string) (*html.Node, error) {
	reader := bytes.NewReader([]byte(data))

	parent := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}

	fragments, err := html.ParseFragmentWithOptions(reader, parent)

	if err != nil {
		return nil, err
	}

	for _, node := range fragments {
		parent.AppendChild(node)
	}

	return parent, nil
}

// renderNodeToString todo
func renderNodeToString(e *html.Node) (string, error) {
	var b bytes.Buffer
	err := html.Render(&b, e)
	return b.String(), err
}

// renderNodesToString todo
func renderNodesToString(nodes []*html.Node) (string, error) {
	text := ""

	for _, node := range nodes {
		rendered, err := renderNodeToString(node)

		if err != nil {
			return "", err
		}

		text += rendered
	}

	return text, nil
}

func renderInnerHTML(parent *html.Node) (string, error) {
	return renderNodesToString(nodeChildren(parent))
}

func selfIndexOfNode(n *html.Node) int {
	ix := 0
	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		ix++
	}

	return ix
}

func isChildrenTheSame(actual, proposed *html.Node) bool {
	proposedString, err := renderNodesToString(nodeChildren(proposed))
	if err != nil {
		return false
	}

	actualString, err := renderNodesToString(nodeChildren(actual))
	if err != nil {
		return false
	}

	return actualString == proposedString
}

func getAllChildrenRecursive(n *html.Node) []*html.Node {
	result := make([]*html.Node, 0)

	if n == nil {
		return result
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, c)
		result = append(result, getAllChildrenRecursive(c)...)
	}

	return result
}

// nodeChildren todo
func nodeChildren(n *html.Node) []*html.Node {
	children := make([]*html.Node, 0)

	if n == nil || n.FirstChild == nil {
		return children
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}

	return children
}

func signLiveUIToSelector(e *html.Node, selector *DOMElemSelector) bool {
	if goLiveUidAttr := getAttribute(e, "go-live-uid"); goLiveUidAttr != nil {
		selector.addAttr("go-live-uid", goLiveUidAttr.Val)

		if keyAttr := getAttribute(e, "key"); keyAttr != nil {
			selector.addAttr("key", keyAttr.Val)
		}
		return true
	}
	return false
}

// selectorFromNode
func selectorFromNode(e *html.Node) (*DOMSelector, error) {

	if e == nil {
		return nil, ErrComponentNil
	}

	selector := NewDOMSelector()

	for parent := e; parent != nil; parent = parent.Parent {

		es := NewDOMElementSelector()
		es.setElemen("*")

		if signLiveUIToSelector(parent, es) {
			selector.addParentSelector(es)
		} else {
			return nil, ErrElementNotSigned
		}

		if goLiveComponentIDAttr := getAttribute(parent, ComponentIdAttrKey); goLiveComponentIDAttr != nil {
			es.addAttr(ComponentIdAttrKey, goLiveComponentIDAttr.Val)
			return selector, nil
		}
	}

	return nil, ErrCouldNotProvideValidSelector
}

// pathToComponentRoot todo
func pathToComponentRoot(e *html.Node) []int {

	path := make([]int, 0)

	for parent := e; parent != nil; parent = parent.Parent {

		attrs := AttrMapFromNode(parent)

		path = append([]int{selfIndexOfNode(parent)}, path...)

		if _, ok := attrs[ComponentIdAttrKey]; ok {
			return path
		}
	}

	return path
}

func removeNodeAttribute(e *html.Node, key string) {
	n := make([]html.Attribute, 0)

	for _, attr := range e.Attr {
		if attr.Key == key {
			continue
		}
		n = append(n, attr)
	}

	e.Attr = n
}

func addNodeAttribute(e *html.Node, key, value string) {
	e.Attr = append(e.Attr, html.Attribute{
		Key: key,
		Val: value,
	})
}

func getAttribute2(e *html.Node, key string) (*html.Attribute, bool) {
	attr := getAttribute(e, key)

	return attr, attr != nil
}

func getAttribute(e *html.Node, key string) *html.Attribute {
	for _, attr := range e.Attr {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}
