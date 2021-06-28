package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"strings"
)

func main() {

	baseHtml, _ := ioutil.ReadFile("./ci/base.html")
	mainJs, _ := ioutil.ReadFile("./ci/main.js")

	//m := minify.New()

	//m.AddFunc("application/javascript", js.Minify)
	//minifiedMainJs, _ := m.Bytes("application/javascript", mainJs)

	finalHtml := strings.Replace(string(baseHtml), "<!-- script:main -->", string(mainJs), 1)

	err := ioutil.WriteFile("./live/page.html", []byte(finalHtml), fs.ModePerm)

	fmt.Println(err)

}
