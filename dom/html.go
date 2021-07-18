package dom

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	ErrCouldNotProvideValidSelector = fmt.Errorf("could not provide a valid selector")
	ErrElementNotSigned             = fmt.Errorf("element is not signed with gl-uid")
	ErrElementNotFound              = fmt.Errorf("element not found")
)

const ComponentIdAttrKey = "gl-cid"
const ElementUidAttrKey = "gl-uid"
const ElementKeyAttrKey = "key"

// AttrMapFromNode todo
func AttrMapFromNode(node *html.Node) map[string]string {
	m := map[string]string{}
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}

// nodeFromString todo
func NodeFromString(data string) (*html.Node, error) {
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

// RenderNodeToString todo
func RenderNodeToString(e *html.Node) (string, error) {
	var b bytes.Buffer
	err := html.Render(&b, e)
	return b.String(), err
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

func RenderInnerHTML(parent *html.Node) (string, error) {
	return RenderNodesToString(nodeChildren(parent))
}

func SelectorFromNode(e *html.Node) (*DomSelector, error) {

	if e == nil {
		return nil, ErrElementNotFound
	}

	selector := newDomSelector()

	for ; e != nil; e = e.Parent {

		es := newDOMElementSelector()
		es.setElement(e.DataAtom.String())

		if !selector.HasAttr(ElementUidAttrKey) {
			goLiveUidAttr, found := GetLiveUidAttributeValue(e)

			if found {
				es.addAttr(ElementUidAttrKey, goLiveUidAttr)
			}
		}

		if !selector.HasAttr(ElementKeyAttrKey) {
			if keyAttr := GetAttribute(e, ElementKeyAttrKey); keyAttr != nil {
				es.addAttr("key", keyAttr.Val)
			}
		}

		//if !selector.HasAttr(ComponentIdAttrKey) {
		//	if goLiveComponentIDAttr := GetAttribute(e, ComponentIdAttrKey); goLiveComponentIDAttr != nil {
		//		es.addAttr(ComponentIdAttrKey, goLiveComponentIDAttr.Val)
		//	}
		//}

		if len(es.query) > 1 {
			selector.addParentSelector(es)
		}
	}

	return selector, nil
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

func GetAllChildrenRecursive(n *html.Node, name string) []*html.Node {
	result := make([]*html.Node, 0)

	if n == nil {
		return result
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, c)

		cid := GetAttribute(c, "gl-cid")

		if cid != nil && cid.Val != name {
			continue
		}

		result = append(result, GetAllChildrenRecursive(c, name)...)
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

// pathToComponentRoot todo
func PathToComponentRoot(e *html.Node) []int {

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

func RemoveNodeAttribute(e *html.Node, key string) {
	n := make([]html.Attribute, 0)

	for _, attr := range e.Attr {
		if attr.Key == key {
			continue
		}
		n = append(n, attr)
	}

	e.Attr = n
}

func AddNodeAttribute(e *html.Node, key, value string) {
	e.Attr = append(e.Attr, html.Attribute{
		Key: key,
		Val: value,
	})
}

func GetLiveUidAttributeValue(e *html.Node) (string, bool) {
	a := GetAttribute(e, ElementUidAttrKey)

	if a == nil {
		return "", false
	}

	return a.Val, true
}

func GetAttribute(e *html.Node, key string) *html.Attribute {
	for _, attr := range e.Attr {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}

func NodeIsText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func NodeIsElement(node *html.Node) bool {
	return node != nil && node.Type == html.ElementNode
}

func NextRelevantElement(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}

	for node = node.NextSibling; !NodeRelevant(node) && node != nil; node = node.NextSibling {

	}

	return node
}

func NodeRelevant(node *html.Node) bool {
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

	aSelector, err := SelectorFromNode(a)

	if err != nil || aSelector == nil {
		return false
	}

	bSelector, err := SelectorFromNode(b)

	if err != nil || bSelector == nil {
		return false
	}

	return aSelector.ToString() == bSelector.ToString()
}
