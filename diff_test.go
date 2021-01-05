package golive

import (
	"reflect"
	"testing"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	a := `<body go-live-component-id><h1>Hello world<span>a</span></h1></body>`
	b := `<body go-live-component-id><h1>Hello world<span></span></h1></body>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	if len(diffs) != 1 {
		t.Error("expecting diff to have length of 1, received", len(diffs))
		return
	}

	diff := diffs[0]
	if diff.Type != SetInnerHtml {
		t.Error("err, expecting type =", SetInnerHtml, "received =", diff.Type)
	}
}

func TestDiffPathFromRemovedNode(t *testing.T) {
	t.Parallel()

	a := `<body go-live-component-id><h1>Hello world<span>a</span></h1></body>`
	b := `<body go-live-component-id><h1>Hello world<span></span></h1></body>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	if len(diffs) != 1 {
		t.Error("expecting 1 diff")
	}

	if diffs[0].Type != SetInnerHtml {
		t.Error("expecting diff to be of type set inner html")
	}
}

func TestDiffPathFromAppendTextAndSpan(t *testing.T) {
	t.Parallel()

	a := `<div go-live-component-id><h1>Hello world<span>a</span></h1></div>`
	b := `<div go-live-component-id><h1>Hello world<span></span><span></span></h1></div>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	if len(diffs) != 2 {
		t.Error("expecting to have 2 diffs receiving", len(diffs))
		return
	}

	for _, diff := range diffs {
		selector := PathToComponentRoot(diff.Element)

		if diff.Type == Append {
			if !reflect.DeepEqual(selector, []int{0, 0}) {
				t.Error("err, unexpected selector =", selector)
				return
			}
		} else if diff.Type == SetInnerHtml {
			if !reflect.DeepEqual(selector, []int{0, 0, 1}) {
				t.Error("err, unexpected selector =", selector)
				return
			}
		} else {
			t.Error("err, unexpected type =", diff.Type)
			return
		}
	}
}

func TestGetDiffFromNodes(t *testing.T) {

	t.Parallel()
	a := `<div go-live-component-id><span></span></div>`
	b := `<div go-live-component-id><span>1</span></div>`

	diffs, _ := GetDiffFromRawHTML(a, b)

	if len(diffs) != 1 {
		t.Error("unexpected length")
	}

	diff := diffs[0]

	if diff.Content != "1" {
		t.Error("unexpected content")
	}

	if diff.Type != SetInnerHtml {
		t.Error("unexpected operation type, expecting append")
	}
}

func TestDiffRegressionToLongerThanFrom(t *testing.T) {
	t.Parallel()

	a := `<div go-live-component-id><h1>Hello world<span></span><span></span></h1></div>`
	b := `<div go-live-component-id><h1>Hello world<span>a</span></h1></div>`

	diffs, err := GetDiffFromRawHTML(a, b)

	if len(diffs) != 2 {
		t.Error("expecting 2 diffs receiving", len(diffs))
		return
	}

	for _, diff := range diffs {
		selector := PathToComponentRoot(diff.Element)
		if diff.Type == Remove {
			if !reflect.DeepEqual(selector, []int{0, 0, 2}) {
				t.Error("err, unexpected selector =", selector)
			}
		} else if diff.Type == SetInnerHtml {
			if diff.Content != "a" {
				t.Error("content is wrong")
			}

			if !reflect.DeepEqual(selector, []int{0, 0, 1}) {
				t.Error("err, unexpected selector =", selector)
			}
		} else {
			t.Error("err, unexpected type =", diff.Type)
		}
	}

	if err != nil {
		t.Fatal(err)
	}
}
