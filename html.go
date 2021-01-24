package golive

import (
	"bytes"
	"fmt"
	"strings"

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

func selfIndexOfElement(n *html.Node) int {
	ix := 0
	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		if prev.Type == html.ElementNode {
			ix++
		}
	}

	return ix
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

func nodeChildrenElements(n *html.Node) []*html.Node {
	children := make([]*html.Node, 0)

	if n == nil || n.FirstChild == nil {
		return children
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		children = append(children, child)
	}

	return children
}

func signLiveUIToSelector(e *html.Node, selector *domElemSelector) bool {
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
func selectorFromNode(e *html.Node) (*domSelector, error) {

	if e == nil {
		return nil, ErrComponentNil
	}

	selector := newDomSelector()

	for parent := e; parent != nil; parent = parent.Parent {

		es := newDOMElementSelector()
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

func getLiveUidAttributeValue(e *html.Node) (string, bool) {
	a := getAttribute(e, "go-live-uid")

	if a == nil {
		return "", false
	}

	return a.Val, true
}

func getAttribute(e *html.Node, key string) *html.Attribute {
	for _, attr := range e.Attr {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}

func nodeIsText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}
func nodeIsElement(node *html.Node) bool {
	return node != nil && node.Type == html.ElementNode
}

func nextRelevantElement(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}

	for node = node.NextSibling; !nodeRelevant(node) && node != nil; node = node.NextSibling {

	}

	return node
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
