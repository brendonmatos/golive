package golive

import "fmt"

// LiveComponentWrapper is a struct
type LiveComponentWrapper struct {
	Name      string
	lifeCycle *ComponentLifeCycle
	component *LiveComponent
}

func (l *LiveComponentWrapper) Prepare(lc *LiveComponent) {
	l.lifeCycle = lc.updatesChannel
	l.component = lc
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
	if !l.component.IsMounted {
		fmt.Printf("Not mounted")
	}

	*l.lifeCycle <- ComponentLifeTimeMessage{
		Stage:     Updated,
		Component: l.component,
	}
}
