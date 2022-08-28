package dom

import "strings"

type Selector struct {
	query []*domElemSelector
}

func newDomSelector() *Selector {
	return &Selector{
		query: make([]*domElemSelector, 0),
	}
}

func (ds *Selector) addChild() *domElemSelector {
	de := newDOMElementSelector()
	ds.addChildSelector(de)
	return de
}

func (ds *Selector) addParentSelector(d *domElemSelector) {
	ds.query = append([]*domElemSelector{d}, ds.query...)
}

func (ds *Selector) addChildSelector(d *domElemSelector) {
	ds.query = append(ds.query, d)
}
func (ds *Selector) addParent() *domElemSelector {
	de := newDOMElementSelector()
	ds.addParentSelector(de)
	return de
}

func (ds *Selector) ToString() string {
	var e []string

	for _, q := range ds.query {
		e = append(e, q.ToString())
	}

	return strings.TrimSpace(strings.Join(e, " "))
}

func (ds *Selector) HasAttr(key string) bool {
	for _, q := range ds.query {
		if q.HasAttr(key) {
			return true
		}
	}

	return false
}

type domElemSelector struct {
	query []string
}

func newDOMElementSelector() *domElemSelector {
	return &domElemSelector{
		query: []string{},
	}
}

func (de *domElemSelector) setElement(element string) {
	de.query = append(de.query, element)
}

func (de *domElemSelector) addAttr(key, value string) {
	de.query = append(de.query, "[", key, "=\"", value, "\"]")
}

func (de *domElemSelector) ToString() string {
	return strings.Join(de.query, "")
}

func (de *domElemSelector) IsEmpty() bool {
	return len(de.query) == 0
}

func (de *domElemSelector) HasAttr(attr string) bool {
	return strings.Contains(de.ToString(), attr)
}
