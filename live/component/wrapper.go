package component

import (
	"github.com/brendonmatos/golive"
)

// Wrapper is a struct
type Wrapper struct {
	Name      string
	Component *Component
}

func (l *Wrapper) Create(lc *Component) {
	l.Component = lc
}

// TemplateHandler ...
func (l *Wrapper) TemplateHandler(_ *Component) string {
	return "<div></div>"
}

// BeforeMount the Component loading html
func (l *Wrapper) BeforeMount(_ *Component) {
}

// Mounted the Component loading html
func (l *Wrapper) Mounted(_ *Component) {
}

// BeforeUnmount before we kill the Component
func (l *Wrapper) BeforeUnmount(_ *Component) {
}

// Commit puts an boolean to the commit channel and notifies who is listening
func (l *Wrapper) Commit() {
	l.Component.Log(golive.LogTrace, "Updated", golive.LogEx{"name": l.Component.Name})
	l.Component.Update()
}
