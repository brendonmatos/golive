package differ

import (
	"golang.org/x/net/html"
	"strconv"
)

type Type int

func (dt Type) ToString() string {
	return strconv.Itoa(int(dt))
}

const (
	Append Type = iota
	Remove
	SetInnerHTML
	SetAttr
	RemoveAttr
	Replace
	Move
)

type ChangeInstruction struct {
	ChangeType Type
	Element    *html.Node
	Content    string
	Attr       AttrChange
	Index      int
}

// AttrChange todo
type AttrChange struct {
	Name  string
	Value string
}

type Diff struct {
	Actual       *html.Node
	Instructions []ChangeInstruction
	quantity     int
	doneElements []*html.Node
}

func NewDiff(actual *html.Node) *Diff {
	return &Diff{
		Actual:       actual,
		Instructions: make([]ChangeInstruction, 0),
	}
}

func (d *Diff) InstructionsByType(t Type) []ChangeInstruction {
	s := make([]ChangeInstruction, 0)

	for _, i := range d.Instructions {
		if i.ChangeType == t {
			s = append(s, i)
		}
	}

	return s
}

// changed since last checkpoint
func (d *Diff) changed() bool {
	return len(d.Instructions) != d.quantity
}

func (d *Diff) Propose(proposed *html.Node) {
	d.clearMarked()
	d.diffNode(d.Actual, proposed)
}

func (d *Diff) checkpoint() {
	d.quantity = len(d.Instructions)
}

func (d *Diff) diffNode(actual, proposed *html.Node) {

	if actual == nil || proposed == nil {
		d.diffWalk(actual, proposed)
		return
	}

	uidActual, actualOk := getLiveUidAttributeValue(actual)
	uidProposed, proposedOk := getLiveUidAttributeValue(proposed)

	if actualOk && proposedOk && uidActual != uidProposed {
		content, _ := RenderNodeToString(proposed)
		d.Instructions = append(d.Instructions, ChangeInstruction{
			ChangeType: Replace,
			Element:    actual,
			Content:    content,
		})
		return
	}

	d.diffNodeAttributes(actual, proposed)
	d.diffWalk(actual.FirstChild, proposed.FirstChild)
	d.markNodeDone(proposed)
}

func (d *Diff) clearMarked() {
	d.doneElements = make([]*html.Node, 0)
}

func (d *Diff) markNodeDone(node *html.Node) {
	d.doneElements = append(d.doneElements, node)
}

func (d *Diff) isMarked(node *html.Node) bool {
	for _, n := range d.doneElements {
		if n == node {
			return true
		}
	}

	return false
}

func (d *Diff) diffWalk(actual, proposed *html.Node) {

	if actual == nil && proposed == nil {
		return
	}

	if nodeIsText(actual) || nodeIsText(proposed) {
		d.checkpoint()
		d.diffTextNode(actual, proposed)
		if d.changed() {
			return
		}
	}

	if actual != nil && proposed != nil {
		d.diffNode(actual, proposed)
		goto next
	}

	if actual == nil && proposed != nil && nodeIsElement(proposed) {
		nodeContent, _ := RenderNodeToString(proposed)

		d.Instructions = append(d.Instructions, ChangeInstruction{
			ChangeType: Append,
			Element:    proposed.Parent,
			Content:    nodeContent,
		})
		d.markNodeDone(proposed)

		goto next
	}

	if proposed == nil && actual != nil && nodeIsElement(actual) {
		d.Instructions = append(d.Instructions, ChangeInstruction{
			ChangeType: Remove,
			Element:    actual,
		})
		d.markNodeDone(actual)

		goto next
	}

next:
	nextActual := nextRelevantElement(actual)
	nextProposed := nextRelevantElement(proposed)

	/**
	esse check serve para verificar
	se alguma operacao ainda 'e necessaria.
	Se os dois serem nulos 'e pq a branch
	atual nao tem mais nada para comparar
	seria o final do conteudo comparavel.
	*/
	if nextActual != nil || nextProposed != nil {
		d.diffWalk(nextActual, nextProposed)
	}

}

func (d *Diff) forceRenderElementContent(proposed *html.Node) {
	childrenHTML, _ := RenderInnerHTML(proposed)

	d.Instructions = append(d.Instructions, ChangeInstruction{
		ChangeType: SetInnerHTML,
		Content:    childrenHTML,
		Element:    proposed,
	})
}

// diffNodeAttributes compares the attributes of el with those of otherEl
// and adds the necessary patches to make the attributes in el match those in
// otherEl
func (d *Diff) diffNodeAttributes(actual, proposed *html.Node) {

	actualAttrs := AttrMapFromNode(actual)
	proposedAttrs := AttrMapFromNode(proposed)

	// Now iterate through the attributes in otherEl
	for name, otherValue := range proposedAttrs {
		value, found := actualAttrs[name]
		if !found || value != otherValue {
			d.Instructions = append(d.Instructions, ChangeInstruction{
				ChangeType: SetAttr,
				Element:    actual,
				Attr: AttrChange{
					Name:  name,
					Value: otherValue,
				},
			})
		}
	}

	for attrName := range actualAttrs {
		if _, found := proposedAttrs[attrName]; !found {

			d.Instructions = append(d.Instructions, ChangeInstruction{
				ChangeType: RemoveAttr,
				Element:    actual,
				Attr: AttrChange{
					Name: attrName,
				},
			})
		}
	}
}

func (d *Diff) diffTextNode(actual, proposed *html.Node) {

	// Any node is text
	if !nodeIsText(proposed) && !nodeIsText(actual) {
		return
	}

	proposedIsRelevant := nodeRelevant(proposed)
	actualIsRelevant := nodeRelevant(actual)

	if !proposedIsRelevant && !actualIsRelevant {
		return
	}

	// XOR
	if proposedIsRelevant != actualIsRelevant {
		goto renderEntireNode
	}

	if proposed.Data != actual.Data {
		goto renderEntireNode
	}

	return

renderEntireNode:
	{

		node := proposed

		if node == nil {
			node = actual
		}

		if node == nil {
			return
		}

		d.forceRenderElementContent(node.Parent)
		d.markNodeDone(node.Parent)
	}

}
