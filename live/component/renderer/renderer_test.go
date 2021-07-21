package renderer

import (
	"fmt"
	"testing"
)

func TestSignHtmlString(t *testing.T) {

	s := signHtmlTemplate(`
		<div gl-input="Input">
			%d
		</div>
	`, "123")

	fmt.Println(s)

}
