package golive

import (
	"reflect"
	"regexp"
	"runtime/debug"
	"testing"
	"time"
)

type diffTest struct {
	template  string
	diff      *Diff
	component *LiveComponent
}

type diffComponent struct {
	LiveComponentWrapper
	testTemplate string
	Check        bool
}

var reSelectGoliveAttr = regexp.MustCompile(`[ ]?go-live-uid="[a-zA-Z0-9_\-]+"`)

func (l *diffComponent) TemplateHandler(lc *LiveComponent) string {
	return l.testTemplate
}

func newDiffTest(d diffTest) diffTest {
	dc := diffComponent{}

	c := NewLiveComponent("testcomp", &dc)

	d.component = c
	c.log = NewLoggerBasic().Log

	dc.testTemplate = d.template

	_ = c.Create(nil)
	_ = c.Mount()

	_, _ = c.Render()

	dc.Check = true

	df, _ := c.LiveRender()

	d.diff = df

	return d
}

func (d *diffTest) assert(expectations []changeInstruction, t *testing.T) {

	if len(d.diff.instructions) != len(expectations) {
		t.Error("The number of instruction are len", len(d.diff.instructions), "expected to be len", len(expectations))
	}

	for indexExpected, expected := range expectations {

		foundGiven := false

		for indexGiven, given := range d.diff.instructions {

			if indexExpected == indexGiven {
				foundGiven = true
			} else {
				continue
			}

			if given.changeType != expected.changeType {
				t.Error("type is different given:", given.changeType, "expeted:", expected.changeType)
			}
			a := reSelectGoliveAttr.ReplaceAllString(given.content, "")

			if a != expected.content {
				t.Error("contents are different given:", a, "expeted:", expected.content)
			}

			if expected.attr != (attrChange{}) && !reflect.DeepEqual(given.attr, expected.attr) {
				t.Error("attributes are different given:", given.attr, "expeted:", expected.attr)
			}

			if !reflect.DeepEqual(pathToComponentRoot(given.element), pathToComponentRoot(expected.element)) {
				t.Error("elements with different elements given:", pathToComponentRoot(given.element), "expeted:", pathToComponentRoot(expected.element))
			}

		}

		if !foundGiven {
			t.Error("given instruction not found")
		}

	}

	if t.Failed() {
		t.Log("Time", time.Now().Format(time.Kitchen), string(debug.Stack()))
	}
}

func TestDiff_RemovedNestedText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<h1><span>{{ if .Check }}{{else}}hello world{{ end }}</span></h1>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild.LastChild,
			content:    "",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ChangeNestedText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>Hello world<span>{{ if .Check }}hello{{ else }}hello world{{ end }}</span></div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild.LastChild,
			content:    "hello",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_RemoveElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}{{else}}<div></div>{{ end }}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: Remove,
			element:    dt.diff.actual.FirstChild.FirstChild,
			content:    "",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div></div>{{else}}{{ end }}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: Append,
			element:    dt.diff.actual.FirstChild,
			content:    "<div></div>",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendNestedElements(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div><div></div></div>{{ end }}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: Append,
			element:    dt.diff.actual.FirstChild,
			content:    "<div><div></div></div>",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceNestedElementsWithText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div>a<div>a</div></div>{{ else }}<span></span>{{end}}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.FirstChild,
			content:    "<div>a<div>a</div></div>",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceTagWithContent(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div>a</div>{{ else }}<span>a</span>{{ end }}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.FirstChild,
			content:    "<div>a</div>",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AddAttribute(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div {{ if .Check }}disabled{{ end }}></div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			content:    "",
			attr: attrChange{
				name:  "disabled",
				value: "",
			},
		},
	}, t)
}

func TestDiff_RemoveAttribute(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div {{ if not .Check }}disabled{{ end }}></div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: RemoveAttr,
			element:    dt.diff.actual.FirstChild,
			content:    "",
			attr: attrChange{
				name:  "disabled",
				value: "",
			},
		},
	}, t)
}

func TestDiff_AddTextContent(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}aaaa{{ end }}</div>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild,
			content:    "aaaa",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffWithTabs(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `{{ if .Check }}
	<div></div>
{{else}}
<div></div>
{{end}}`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			content:    "",
			attr:       attrChange{},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			content:    "",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffWithTabsAndBreakLine(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `{{ if .Check }}
	<div></div>
{{else}}

<div></div>
{{end}}`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			content:    "",
			attr:       attrChange{},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			content:    "",
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffAttr(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<button {{if .Check}}disabled="disabled"{{end}}></button>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
	}, t)
}

func TestDiff_DiffAttrs(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
	}, t)
}

func TestDiff_DiffMultiElementAndAttrs(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button><button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button>`,
	})

	dt.assert([]changeInstruction{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild.NextSibling,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
	}, t)
}
