package golive

import "strings"

type domSelector struct {
	query []*domElemSelector
}

func newDomSelector() *domSelector {
	return &domSelector{
		query: make([]*domElemSelector, 0),
	}
}

func (ds *domSelector) addChild() *domElemSelector {
	de := newDOMElementSelector()
	ds.addChildSelector(de)
	return de
}

func (ds *domSelector) addParentSelector(d *domElemSelector) {
	ds.query = append([]*domElemSelector{d}, ds.query...)
}

func (ds *domSelector) addChildSelector(d *domElemSelector) {
	ds.query = append(ds.query, d)
}
func (ds *domSelector) addParent() *domElemSelector {
	de := newDOMElementSelector()
	ds.addParentSelector(de)
	return de
}

func (ds *domSelector) toString() string {
	var e []string

	for _, q := range ds.query {
		e = append(e, q.toString())
	}

	return strings.Join(e, " ")
}

type domElemSelector struct {
	query []string
}

func newDOMElementSelector() *domElemSelector {
	return &domElemSelector{
		query: []string{},
	}
}

func (de *domElemSelector) setElemen(elemn string) {
	de.query = append(de.query, elemn)
}

func (de *domElemSelector) addAttr(key, value string) {
	de.query = append(de.query, "[", key, "=\"", value, "\"]")
}
func (de *domElemSelector) toString() string {
	return strings.Join(de.query, "")
}
