package golive

import "strings"

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
