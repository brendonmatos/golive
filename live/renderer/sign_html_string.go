package renderer

import (
	"github.com/brendonmatos/golive/dom"
	"github.com/brendonmatos/golive/live/util"
	"regexp"
	"strings"
)

func replaceWithFunction(content string, r *regexp.Regexp, h func(string) string) string {
	matches := r.FindAllStringSubmatchIndex(content, -1)

	util.ReverseSlice(matches)

	for _, match := range matches {
		startIndex := match[0]
		endIndex := match[1]

		startSlice := content[:startIndex]
		endSlide := content[endIndex:]
		matchedSlice := content[startIndex:endIndex]

		content = startSlice + h(matchedSlice) + endSlide
	}

	return content
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+[ ]?)`)

func signHtmlTemplate(template string, uid string) string {

	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` ` + dom.ComponentIdAttrKey + `="` + uid + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}

	template = replaceWithFunction(template, rxTagName, func(s string) string {
		lUid := uid + "_" + util.RandomSmall()
		return s + ` ` + GoLiveUidAttrKey + `="` + lUid + `" `
	})

	return template
}
