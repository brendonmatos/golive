package renderer

import (
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/util"
	"strings"
)

func signHtmlString(template string, uid string) string {
	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` ` + differ.ComponentIdAttrKey + `="` + uid + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}

	matches := rxTagName.FindAllStringSubmatchIndex(template, -1)

	util.ReverseSlice(matches)

	for _, match := range matches {
		startIndex := match[0]
		endIndex := match[1]

		startSlice := template[:startIndex]
		endSlide := template[endIndex:]
		matchedSlice := template[startIndex:endIndex]

		lUid := uid + "_" + util.RandomSmall()
		replaceWith := matchedSlice + ` ` + GoLiveUidAttrKey + `="` + lUid + `" `
		template = startSlice + replaceWith + endSlide
	}

	//t, _ := template.New(c.Name).Funcs(template.FuncMap{
	//	"render": c.RenderChild,
	//}).Parse(tr.templateString)

	//tr.template = t

	return template

}
