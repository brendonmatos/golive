package main

import (
	"io/ioutil"
	"strings"
)

func main() {

	read, _ := ioutil.ReadFile("./ci/base.html")
	base := string(read)

	code := []string{"var BasePageString = `", base, "`"}
	contents := []string{"package golive", strings.Join(code, "")}

	ioutil.WriteFile("html_page.go", []byte(strings.Join(contents, "\n")), 0)
}
