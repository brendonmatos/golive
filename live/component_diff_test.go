package live

import (
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"reflect"
	"regexp"
	"runtime/debug"
	"testing"
	"time"

	"golang.org/x/net/html"
)

type diffTest struct {
	template  string
	diff      *differ.Diff
	component *Component
}

type instructionExpect struct {
	changeType differ.Type
	element    *html.Node
	content    *string
	attr       differ.AttrChange
	index      *int
}

type diffComponent struct {
	Wrapper
	testTemplate string
	Check        bool
}

var reSelectGoliveAttr = regexp.MustCompile(`[ ]?gl-uid="[a-zA-Z0-9_\-]+"`)

func (l *diffComponent) TemplateHandler(_ *Component) string {
	return l.testTemplate
}

func newDiffTest(d diffTest) diffTest {
	dc := diffComponent{}

	c := DefineComponent("testcomp")

	c.SetState(dc)

	c.Log = golive.NewLoggerBasic().Log

	dc.testTemplate = d.template

	err := c.Mount()
	if err != nil {
		panic(err)
	}

	dc.Check = true

	df, _ := c.LiveRender()

	d.diff = df

	return d
}

func (d *diffTest) assert(expectations []instructionExpect, t *testing.T) {

	if len(d.diff.Instructions) != len(expectations) {
		t.Error("The number of instruction are len", len(d.diff.Instructions), "expected to be len", len(expectations))
	}

	for indexExpected, expected := range expectations {

		foundGiven := false

		for indexGiven, given := range d.diff.Instructions {

			if indexExpected == indexGiven {
				foundGiven = true
			} else {
				continue
			}

			if given.ChangeType != expected.changeType {
				t.Error("type is different given:", given.ChangeType, "expeted:", expected.changeType)
			}
			a := reSelectGoliveAttr.ReplaceAllString(given.Content, "")

			if expected.content != nil && a != *expected.content {
				t.Error("contents are different given:", a, "expeted:", *expected.content)
			}

			if expected.attr != (differ.AttrChange{}) && !reflect.DeepEqual(given.Attr, expected.attr) {
				t.Error("attributes are different given:", given.Attr, "expeted:", expected.attr)
			}

			if !reflect.DeepEqual(differ.PathToComponentRoot(given.Element), differ.PathToComponentRoot(expected.element)) {
				t.Error("elements with different elements given:", differ.PathToComponentRoot(given.Element), "expeted:", differ.PathToComponentRoot(expected.element))
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
			changeType: differ.SetInnerHTML,
			element:    dt.diff.Actual.FirstChild.LastChild,
			attr:       differ.AttrChange{},
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
			changeType: differ.SetInnerHTML,
			element:    dt.diff.Actual.FirstChild.LastChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.Remove,
			element:    dt.diff.Actual.FirstChild.FirstChild,
			attr:       differ.AttrChange{},
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
			changeType: differ.Append,
			element:    dt.diff.Actual.FirstChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.Append,
			element:    dt.diff.Actual.FirstChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.Replace,
			element:    dt.diff.Actual.FirstChild.FirstChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.Replace,
			element:    dt.diff.Actual.FirstChild.FirstChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "",
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
			changeType: differ.RemoveAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "",
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
			changeType: differ.SetInnerHTML,
			element:    dt.diff.Actual.FirstChild,
			content:    &c,
			attr:       differ.AttrChange{},
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
			changeType: differ.Replace,
			element:    dt.diff.Actual.FirstChild.NextSibling,
			attr:       differ.AttrChange{},
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
			changeType: differ.Replace,
			element:    dt.diff.Actual.FirstChild.NextSibling,
			attr:       differ.AttrChange{},
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
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "disabled",
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
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "disabled",
			},
		},
		{
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "class",
				Value: "hello world",
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
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "disabled",
			},
		},
		{
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild,
			attr: differ.AttrChange{
				Name:  "class",
				Value: "hello world",
			},
		},
		{
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild.NextSibling,
			attr: differ.AttrChange{
				Name:  "disabled",
				Value: "disabled",
			},
		},
		{
			changeType: differ.SetAttr,
			element:    dt.diff.Actual.FirstChild.NextSibling,
			attr: differ.AttrChange{
				Name:  "class",
				Value: "hello world",
			},
		},
	}, t)
}

func TestDiff_DiffMultiKey(t *testing.T) {
	t.Parallel()

	_ = newDiffTest(diffTest{
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

}
