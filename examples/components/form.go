package components

import "github.com/brendonmatos/golive"

type Form struct {
	golive.LiveComponentWrapper
	DynamicInput  *golive.LiveComponent
	InputtedValue *string
}

func NewForm() *golive.LiveComponent {

	var value = ""

	return golive.NewLiveComponent("Form", &Form{
		InputtedValue: &value,
		DynamicInput:  NewDynamicInput(DynamicInputProps{Value: &value}),
	})
}

func (d *Form) TemplateHandler(_ *golive.LiveComponent) string {
	return `<div>
		{{ render .DynamicInput }}

		<div>
			Value inputed: {{ .InputtedValue }}
		</div>
	</div>`
}
