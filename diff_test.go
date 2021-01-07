package golive

import (
	"fmt"
	"testing"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	a, _ := CreateDOMFromString(`<body><h1>Hello world<span>a</span></h1></body>`)
	b, _ := CreateDOMFromString(`<body><h1>Hello world<span></span></h1></body>`)

	diff := NewDiff(a)

	diff.Propose(b)

	fmt.Println(diff.instructions)

}
