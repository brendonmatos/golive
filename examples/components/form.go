package components

import (
	"github.com/brendonmatos/golive/live/component"
)

type Form struct {
	component.Wrapper
	DynamicInput  *component.Component
	InputtedValue *string
}

func NewForm() *component.Component {

	var value = ""

	return component.NewLiveComponent("Form", &Form{
		DynamicInput:  NewDynamicInput(DynamicInputProps{Value: &value}),
		InputtedValue: &value,
	})
}

func (d *Form) TemplateHandler(_ *component.Component) string {
	return `<div>
		{{ render .DynamicInput }}

		<div>
			Value inputed: {{ .InputtedValue }}
		</div>
	</div>`
}
