package golive

import (
	"testing"
)

func TestCreateDOMFromString(t *testing.T) {

	html := `<body><h1>Hello world</h1></body>`

	dom, err := CreateDOMFromString(html)

	if err != nil {
		t.Error(err)
	}

	if dom.Parent != nil {
		t.Error("There is a parent where should not be")
	}
	if dom.FirstChild == nil {
		t.Error("There is not a child where should be")
	}
}

func TestSelectorFromNode(t *testing.T) {
	html := `<body><h1>Hello world<span>a</span></h1></body>`

	dom, _ := CreateDOMFromString(html)

	node := dom.LastChild.LastChild.LastChild.LastChild.FirstChild
	if node.Data != "a" {
		t.Error("value was not parsed correctly")
	}

	if SelectorFromNode(node) != "html:nth-child(1) body:nth-child(2) h1:nth-child(1) span:nth-child(1)" {
		t.Error("wrong selector returned")
	}
}

func TestSelectorFromEmptyNode(t *testing.T) {
	a := `<body><h1>Hello world<span></span></h1></body>`

	dom, _ := CreateDOMFromString(a)

	node := dom.LastChild.LastChild.LastChild.LastChild

	if node.Data != "span" || node.FirstChild != nil {
		t.Error("value was not parsed correctly")
	}

	selector := SelectorFromNode(node)
	if selector != "html:nth-child(1) body:nth-child(2) h1:nth-child(1) span:nth-child(1)" {
		t.Error("wrong selector returned, given", selector)
	}
}
