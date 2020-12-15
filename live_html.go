package golive

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// AttrMapFromNode todo
func AttrMapFromNode(node *html.Node) map[string]string {
	m := map[string]string{}
	for _, attr := range node.Attr {
		m[attr.Key] = attr.Val
	}
	return m
}

// RemoveWhitespaces todo
func RemoveWhitespaces(content string) string {
	return spaceRegex.ReplaceAllString(content, " ")
}

// CreateDOMFromString todo
func CreateDOMFromString(data string) (*html.Node, error) {
	reader := bytes.NewReader([]byte(RemoveWhitespaces(data)))
	return html.Parse(reader)
}

// RenderNodeToString todo
func RenderNodeToString(e *html.Node) string {
	var b bytes.Buffer
	err := html.Render(&b, e)

	if err != nil {
		panic(err)
	}

	return b.String()
}

// RenderNodesToString todo
func RenderNodesToString(nodes []*html.Node) string {
	text := ""

	for _, node := range nodes {
		text += RenderNodeToString(node)
	}

	return text
}

func getClassesSeparated(s string) string {
	return strings.Join(strings.Split(strings.TrimSpace(s), " "), ".")
}

func SelfIndexOfNode(n *html.Node) int {
	ix := 0
	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		if prev.Type == html.TextNode {
			continue
		}

		ix++
	}

	return ix
}

// SelectorFromNode todo
func SelectorFromNode(e *html.Node) string {

	selector := ""
	for parent := e; parent != nil; parent = parent.Parent {

		elementSelector := ""
		attrs := AttrMapFromNode(parent)

		if parent.Type == html.ElementNode {
			elementSelector = parent.Data + elementSelector

			if attr, ok := attrs["id"]; ok {
				if len(attr) > 0 {
					elementSelector = elementSelector + "#" + strings.TrimSpace(attr)
				}
			}

			if attr, ok := attrs["class"]; ok {
				if len(attr) > 0 {
					elementSelector = elementSelector + "." + getClassesSeparated(attr)
				}
			}

			if _, ok := attrs["go-live-component-id"]; ok {
				return selector
			}

			elementSelector = fmt.Sprintf("%s:nth-child(%d)", elementSelector, SelfIndexOfNode(parent)+1)

		}

		if parent.Type == html.TextNode {
			// selector = fmt.Sprintf("*:nth-child(%d) ", SelfIndexOfNode(parent)) + selector
		}

		if len(selector) > 0 {
			selector = elementSelector + " " + selector
		} else {
			selector = elementSelector
		}

	}

	return selector

}
