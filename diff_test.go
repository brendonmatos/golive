package golive

import (
	"testing"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	a := `<body><h1>Hello world<span>a</span></h1></body>`
	b := `<body><h1>Hello world<span></span></h1></body>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	if len(diffs) != 1 {
		t.Error("err")
	}

	diff := diffs[0]
	if diff.Type != SetInnerHtml {
		t.Error("err, expecting type =", SetInnerHtml, "received =", diff.Type)
	}
}

func TestDiff2(t *testing.T) {
	t.Parallel()

	a := `<body><h1>Hello world<span>a</span></h1></body>`
	b := `<body><h1>Hello world<span></span><span></span></h1></body>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	for _, diff := range diffs {
		if diff.Type == Append {
		} else if diff.Type == SetInnerHtml {
		} else {
			t.Error("err, unexpected type =", diff.Type)
		}
	}
}

func TestDiff3(t *testing.T) {
	t.Parallel()

	a := `<body><h1>Hello world<span>a</span></h1></body>`
	b := `<body><h1>Hello world<span></span><span></span></h1></body>`
	diffs, _ := GetDiffFromRawHTML(a, b)

	for _, diff := range diffs {
		if diff.Type == Append {
			if diff.Element != "html:nth-child(1) body:nth-child(2) h1:nth-child(1)" {
				t.Error("err, unexpected type =", diff.Type)
			}
		} else if diff.Type == SetInnerHtml {
			if diff.Element != "html:nth-child(1) body:nth-child(2) h1:nth-child(1) span:nth-child(1)" {
				t.Error("err, unexpected selector =", diff.Element)
			}
		} else {
			t.Error("err, unexpected type =", diff.Type)
		}
	}
}

func TestDiffRegressionToLongerThanFrom(t *testing.T) {
	t.Parallel()

	a := `<body><h1>Hello world<span></span><span></span></h1></body>`
	b := `<body><h1>Hello world<span>a</span></h1></body>`
	_, err := GetDiffFromRawHTML(a, b)

	if err != nil {
		t.Fatal(err)
	}
}
