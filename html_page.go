package golive

import (
	"io/ioutil"
)

var BasePageString string

func init() {
	read, _ := ioutil.ReadFile("./base.html")
	BasePageString = string(read)
}

// var BasePageString = `

// `
