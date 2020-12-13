package golive

import (
	"html/template"
)

// LiveComponentWrapper is a struct
type LiveComponentWrapper struct {
	Name               string
	IsMounted          bool
	HtmlTemplateString string
	HtmlTemplate       *template.Template
	LifeTimeChannel    *LiveTimeChannel
	Rendered           string
}

func (l *LiveComponentWrapper) SetLifeTimeChannel(channel *LiveTimeChannel) {
	l.LifeTimeChannel = channel
}

// TemplateHandler ...
func (l *LiveComponentWrapper) TemplateHandler() string {
	return "<div></div>"
}

// BeforeMount the component loading html
func (l *LiveComponentWrapper) BeforeMount(_ *LiveComponent) {
}

// BeforeMount the component loading html
func (l *LiveComponentWrapper) Mounted(_ *LiveComponent) {
}

// Commit puts an boolean to the commit channel and notifies ho is listening
func (l *LiveComponentWrapper) Commit() {
	*l.LifeTimeChannel <- LifeTimeUpdate
}
