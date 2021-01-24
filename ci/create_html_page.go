package main

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"io/ioutil"
	"strings"
)

func main() {

	baseHtml, _ := ioutil.ReadFile("./ci/base.html")
	mainJs, _ := ioutil.ReadFile("./ci/main.js")

	m := minify.New()

	m.AddFunc("application/javascript", js.Minify)
	minifiedMainJs, _ := m.Bytes("application/javascript", mainJs)

	finalHtml := strings.Replace(string(baseHtml), "<!-- script:main -->", string(minifiedMainJs), 1)

	code := []string{"var BasePageString = `", finalHtml, "`"}
	contents := []string{
		"package golive",
		"// Code automatically generated. DO NOT EDIT.",
		"// > go run ci/create_html_page.go",
		strings.Join(code, ""),
	}

	_ = ioutil.WriteFile("html_page.go", []byte(strings.Join(contents, "\n")), 0)
}
