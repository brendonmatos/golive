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
	LifeTimeChannel    *LifeTimeUpdates
	Rendered           string
}

func (l *LiveComponentWrapper) Prepare(lc *LiveComponent) {
	l.LifeTimeChannel = lc.UpdatesChannel
}

// TemplateHandler ...
func (l *LiveComponentWrapper) TemplateHandler(_ *LiveComponent) string {
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
	*l.LifeTimeChannel <- ComponentLifeTimeMessage{
		Stage:     Updated,
		Component: nil,
	}
}
