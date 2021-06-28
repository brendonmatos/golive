package components

import (
	"github.com/brendonmatos/golive/live"
)

type Form struct {
	live.Wrapper
	DynamicInput  *live.Component
	InputtedValue *string
}

func NewForm() *live.Component {

	var value = ""

	return live.NewLiveComponent("Form", &Form{
		InputtedValue: &value,
		DynamicInput:  NewDynamicInput(DynamicInputProps{Value: &value}),
	})
}

func (d *Form) TemplateHandler(_ *live.Component) string {
	return `<div>
		{{ render .DynamicInput }}

		<div>
			Value inputed: {{ .InputtedValue }}
		</div>
	</div>`
}
