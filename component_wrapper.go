package golive

// LiveComponentWrapper is a struct
type LiveComponentWrapper struct {
	Name      string
	component *LiveComponent
}

func (l *LiveComponentWrapper) Create(lc *LiveComponent) {
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

// BeforeUnmount before we kill the component
func (l *LiveComponentWrapper) BeforeUnmount(_ *LiveComponent) {
}

// Commit puts an boolean to the commit channel and notifies who is listening
func (l *LiveComponentWrapper) Commit() {
	l.component.log(LogTrace, "Updated", logEx{"name": l.component.Name})

	if l.component.life == nil {
		l.component.log(LogError, "call to commit on unmounted component", logEx{"name": l.component.Name})
		return
	}

	l.component.Update()
}
