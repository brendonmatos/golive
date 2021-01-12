package components

import "github.com/brendonmatos/golive"

type Form struct {
	golive.LiveComponentWrapper
	Label        string
	DynamicInput *golive.LiveComponent
}

func NewForm() *golive.LiveComponent {
	return golive.NewLiveComponent("Form", &Form{
		Label:        "",
		DynamicInput: NewDynamicInput(),
	})
}

func (d *Form) TemplateHandler(_ *golive.LiveComponent) string {
	return `<div>
		{{ render .DynamicInput }}
	</div>`
}
