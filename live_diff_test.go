package golive

import "testing"

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

}
