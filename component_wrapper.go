package golive

// LiveComponentWrapper is a struct
type LiveComponentWrapper struct {
	Name      string
	Component *LiveComponent
}

func (l *LiveComponentWrapper) Create(lc *LiveComponent) {
	l.Component = lc
}

// TemplateHandler ...
func (l *LiveComponentWrapper) TemplateHandler(_ *LiveComponent) string {
	return "<div></div>"
}

// BeforeMount the Component loading html
func (l *LiveComponentWrapper) BeforeMount(_ *LiveComponent) {
}

// BeforeMount the Component loading html
func (l *LiveComponentWrapper) Mounted(_ *LiveComponent) {
}

// BeforeUnmount before we kill the Component
func (l *LiveComponentWrapper) BeforeUnmount(_ *LiveComponent) {
}

// Commit puts an boolean to the commit channel and notifies who is listening
func (l *LiveComponentWrapper) Commit() {
	l.Component.log(LogTrace, "Updated", logEx{"name": l.Component.Name})

	if l.Component.life == nil {
		l.Component.log(LogError, "call to commit on unmounted Component", logEx{"name": l.Component.Name})
		return
	}

	l.Component.Update()
}
