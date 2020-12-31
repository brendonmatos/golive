package golive

import (
	"reflect"
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
	html := `<div go-live-component-id><h1>Hello world<span>a</span></h1></div>`

	dom, _ := CreateDOMFromString(html)

	node := dom.LastChild.LastChild.LastChild.LastChild.LastChild.FirstChild
	if node.Data != "a" {
		t.Error("value was not parsed correctly")
	}

	path := PathToComponentRoot(node)
	if !reflect.DeepEqual(path, []int{0, 0, 1, 0}) {
		t.Error("wrong selector returned", path)
	}
}

func TestSelectorFromEmptyNode(t *testing.T) {
	a := `<div go-live-component-id><h1>Hello world<span></span></h1></div>`

	dom, _ := CreateDOMFromString(a)

	node := dom.LastChild.LastChild.LastChild.LastChild.LastChild

	if node.Data != "span" || node.FirstChild != nil {
		t.Error("value was not parsed correctly")
	}

	path := PathToComponentRoot(node)
	if !reflect.DeepEqual(path, []int{0, 0, 1}) {
		t.Error("wrong selector returned", path)
	}
}
