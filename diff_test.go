package golive

import (
	"reflect"
	"runtime/debug"
	"testing"
)

type diffTest struct {
	actual, propose string
	diff            *Diff
}

func newDiffTest(d diffTest) diffTest {
	actual, _ := NodeFromString(d.actual)
	propose, _ := NodeFromString(d.propose)
	d.diff = NewDiff(actual)
	d.diff.Propose(propose)
	return d
}

func (d *diffTest) assert(expectations []ChangeInstruction, t *testing.T) {

	for indexExpected, expected := range expectations {

		foundGiven := false

		for indexGiven, given := range d.diff.instructions {

			if indexExpected == indexGiven {
				foundGiven = true
			} else {
				continue
			}

			if given.Type != expected.Type {
				t.Error("type is different given:", given.Type, "expeted:", expected.Type)
			}

			if given.Content != expected.Content {
				t.Error("contents are different given:", given.Content, "expeted:", expected.Content)
			}

			if !reflect.DeepEqual(given.Attr, expected.Attr) {
				t.Error("attributes are different given:", given.Attr, "expeted:", expected.Attr)
			}

			if !reflect.DeepEqual(PathToComponentRoot(given.Element), PathToComponentRoot(expected.Element)) {
				t.Error("elements with different elements given:", PathToComponentRoot(given.Element), "expeted:", PathToComponentRoot(expected.Element))
			}

		}

		if !foundGiven {
			t.Error("given instruction not found")
		}

	}

	if t.Failed() {
		t.Log(string(debug.Stack()))
	}
}

func TestDiff_RemovedNestedText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<body><h1>Hello world<span>hello world</span></h1></body>`,
		propose: `<body><h1>Hello world<span></span></h1></body>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    SetInnerHtml,
			Element: dt.diff.actual.FirstChild.LastChild,
			Content: "",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_ChangeNestedText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div>Hello world<span>hello world</span></div>`,
		propose: `<div>Hello world<span>hello</span></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    SetInnerHtml,
			Element: dt.diff.actual.FirstChild.LastChild,
			Content: "hello",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_RemoveElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div><div></div></div>`,
		propose: `<div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    Remove,
			Element: dt.diff.actual.FirstChild.FirstChild,
			Content: "",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_AppendElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div></div>`,
		propose: `<div><div></div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    Append,
			Element: dt.diff.actual.FirstChild,
			Content: "<div></div>",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_AppendNestedElements(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div></div>`,
		propose: `<div><div><div></div></div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    Append,
			Element: dt.diff.actual.FirstChild,
			Content: "<div><div></div></div>",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_ReplaceNestedElementsWithText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div><span></span></div>`,
		propose: `<div><div>a<div>a</div></div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    Replace,
			Element: dt.diff.actual.FirstChild.FirstChild,
			Content: "<div>a<div>a</div></div>",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_ReplaceTagWithContent(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div><span>a</span></div>`,
		propose: `<div><div>a</div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    Replace,
			Element: dt.diff.actual.FirstChild.FirstChild,
			Content: "<div>a</div>",
			Attr:    AttrChange{},
		},
	}, t)
}

func TestDiff_AddAttribute(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div></div>`,
		propose: `<div disabled></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    SetAttr,
			Element: dt.diff.actual.FirstChild,
			Content: "",
			Attr: AttrChange{
				Name:  "disabled",
				Value: "",
			},
		},
	}, t)
}

func TestDiff_RemoveAttribute(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div disabled></div>`,
		propose: `<div></div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    RemoveAttr,
			Element: dt.diff.actual.FirstChild,
			Content: "",
			Attr: AttrChange{
				Name:  "disabled",
				Value: "",
			},
		},
	}, t)
}

func TestDiff_AddTextContent(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		actual:  `<div></div>`,
		propose: `<div>aaaa</div>`,
	})

	dt.assert([]ChangeInstruction{
		{
			Type:    SetInnerHtml,
			Element: dt.diff.actual.FirstChild,
			Content: "aaaa",
			Attr:    AttrChange{},
		},
	}, t)
}
