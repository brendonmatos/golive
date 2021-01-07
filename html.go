package golive

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type DOMElemSelector struct {
	query []string
}

func NewDOMElementSelector() *DOMElemSelector {
	return &DOMElemSelector{
		query: []string{},
	}
}

func (de *DOMElemSelector) setElemen(elemn string) {
	de.query = append(de.query, elemn)
}

func (de *DOMElemSelector) addAttr(key, value string) {
	de.query = append(de.query, "[", key, "=\"", value, "\"]")
}
func (de *DOMElemSelector) toString() string {
	return strings.Join(de.query, "")
}

type DOMSelector struct {
	query []*DOMElemSelector
}

func NewDOMSelector() *DOMSelector {
	return &DOMSelector{
		query: make([]*DOMElemSelector, 0),
	}
}

func (ds *DOMSelector) addChild() *DOMElemSelector {
	de := NewDOMElementSelector()
	ds.addChildSelector(de)
	return de
}

func (ds *DOMSelector) addParentSelector(d *DOMElemSelector) {
	ds.query = append([]*DOMElemSelector{d}, ds.query...)
}

func (ds *DOMSelector) addChildSelector(d *DOMElemSelector) {
	ds.query = append(ds.query, d)
}
func (ds *DOMSelector) addParent() *DOMElemSelector {
	de := NewDOMElementSelector()
	ds.addParentSelector(de)
	return de
}

func (ds *DOMSelector) toString() string {
	var e []string

	for _, q := range ds.query {
		e = append(e, q.toString())
	}

	return strings.Join(e, " ")
}

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

func RenderNodeChildren(parent *html.Node) (string, error) {
	return RenderNodesToString(GetChildrenFromNode(parent))
}

func SelfIndexOfNode(n *html.Node) int {
	ix := 0

	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		ix++
	}

	return ix
}

func IsChildrenTheSame(actual, proposed *html.Node) bool {
	proposedString, err := RenderNodesToString(GetChildrenFromNode(proposed))
	if err != nil {
		return false
	}

	actualString, err := RenderNodesToString(GetChildrenFromNode(actual))
	if err != nil {
		return false
	}

	return actualString == proposedString
}

func GetAllChildrenRecursive(n *html.Node) []*html.Node {
	result := make([]*html.Node, 0)

	if n == nil {
		return result
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, c)
		result = append(result, GetAllChildrenRecursive(c)...)
	}

	return result
}

// GetChildrenFromNode todo
func GetChildrenFromNode(n *html.Node) []*html.Node {
	children := make([]*html.Node, 0)

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}

	return children
}

// SelectorFromNode
func SelectorFromNode(e *html.Node) (string, string, error) {

	err := fmt.Errorf("could not provide a valid selector")

	componentId := ""
	selector := NewDOMSelector()

	if e.Type == html.ElementNode {

		attrs := AttrMapFromNode(e)

		es := selector.addChild()
		es.setElemen("*")

		if attr, ok := attrs["go-live-uid"]; ok {
			es.addAttr("go-live-uid", attr)

			if attr, ok := attrs["key"]; ok {
				es.addAttr("key", attr)
			}
		}
	}

	for parent := e.Parent; parent != nil; parent = parent.Parent {

		attrs := AttrMapFromNode(parent)

		es := NewDOMElementSelector()
		es.setElemen("*")

		if attr, ok := attrs["go-live-component-id"]; ok {
			es.addAttr("go-live-component-id", attr)

			componentId = attr

			if attr, ok := attrs["key"]; ok {
				es.addAttr("key", attr)
			}
		}

		if attr, ok := attrs["go-live-uid"]; ok {
			es.addAttr("go-live-uid", attr)

			if attr, ok := attrs["key"]; ok {
				es.addAttr("key", attr)
			}
		}

		selector.addParentSelector(es)

		if componentId == "" {
			continue
		}

		return selector.toString(), componentId, nil
	}

	return "", "", err
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

func getAttribute(e *html.Node, key string) *html.Attribute {
	for _, attr := range e.Attr {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}
