package renderer

import (
	"fmt"
	"testing"
)

func TestSignHtmlTemplate(t *testing.T) {
	s := signHtmlTemplate(`
		<div>
			<input type="checkbox" gl-input="Tasks.2.Done"></input>
		</div>
	`, "aaaaaaa")

	fmt.Println(s)
}
