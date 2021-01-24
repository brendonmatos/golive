package golive

import (
	"fmt"
	"golang.org/x/net/html"
	"reflect"
	"regexp"
	"runtime/debug"
	"testing"
	"time"
)

type diffTest struct {
	template  string
	diff      *diff
	component *LiveComponent
}

type instructionExpect struct {
	changeType DiffType
	element    *html.Node
	content    *string
	attr       attrChange
	index      *int
}

type diffComponent struct {
	LiveComponentWrapper
	testTemplate string
	Check        bool
}

var reSelectGoliveAttr = regexp.MustCompile(`[ ]?go-live-uid="[a-zA-Z0-9_\-]+"`)

func (l *diffComponent) TemplateHandler(_ *LiveComponent) string {
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

func (d *diffTest) assert(expectations []instructionExpect, t *testing.T) {

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

			if expected.content != nil && a != *expected.content {
				t.Error("contents are different given:", a, "expeted:", *expected.content)
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

	dt.assert([]instructionExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild.LastChild,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ChangeNestedText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>Hello world<span>{{ if .Check }}hello{{ else }}hello world{{ end }}</span></div>`,
	})
	c := "hello"
	dt.assert([]instructionExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild.LastChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_RemoveElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}{{else}}<div></div>{{ end }}</div>`,
	})

	dt.assert([]instructionExpect{
		{
			changeType: Remove,
			element:    dt.diff.actual.FirstChild.FirstChild,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendElement(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div></div>{{else}}{{ end }}</div>`,
	})

	c := "<div></div>"
	dt.assert([]instructionExpect{
		{
			changeType: Append,
			element:    dt.diff.actual.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendNestedElements(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div><div></div></div>{{ end }}</div>`,
	})

	c := "<div><div></div></div>"
	dt.assert([]instructionExpect{
		{
			changeType: Append,
			element:    dt.diff.actual.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceNestedElementsWithText(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div>a<div>a</div></div>{{ else }}<span></span>{{end}}</div>`,
	})

	c := "<div>a<div>a</div></div>"
	dt.assert([]instructionExpect{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceTagWithContent(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div>{{ if .Check }}<div>a</div>{{ else }}<span>a</span>{{ end }}</div>`,
	})

	c := "<div>a</div>"
	dt.assert([]instructionExpect{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AddAttribute(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<div {{ if .Check }}disabled{{ end }}></div>`,
	})

	dt.assert([]instructionExpect{
		{
			changeType: SetAttr,
			element:    dt.diff.actual.FirstChild,
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

	dt.assert([]instructionExpect{
		{
			changeType: RemoveAttr,
			element:    dt.diff.actual.FirstChild,
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

	c := "aaaa"
	dt.assert([]instructionExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.diff.actual.FirstChild,
			content:    &c,
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

	dt.assert([]instructionExpect{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.NextSibling,
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

	dt.assert([]instructionExpect{
		{
			changeType: Replace,
			element:    dt.diff.actual.FirstChild.NextSibling,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffAttr(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `<button {{if .Check}}disabled="disabled"{{end}}></button>`,
	})

	dt.assert([]instructionExpect{
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

	dt.assert([]instructionExpect{
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

	dt.assert([]instructionExpect{
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

func TestDiff_DiffMultiKey(t *testing.T) {
	t.Parallel()

	dt := newDiffTest(diffTest{
		template: `
			<div key="1"></div>
			<div key="2"></div>
			{{ if not .Check }}
				<div key="3"></div>
			{{ end }}
			<div key="4">
				<b>Hello world</b>
			</div>
		`,
	})

	fmt.Println(dt.diff.instructions)

}
