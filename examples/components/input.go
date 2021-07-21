package components

import (
	"github.com/brendonmatos/golive/live/component"
)

type DynamicInputProps struct {
	Value *string
	Label string
}

type DynamicInput struct {
	component.Wrapper
	Value *string
	Label string
}

func NewDynamicInput(props DynamicInputProps) *component.Component {
	return component.NewLiveComponent("DynamicInput", &DynamicInput{
		Value: props.Value,
		Label: props.Label,
	})
}

func (d *DynamicInput) TemplateHandler(_ *component.Component) string {
	return `
		<div>
			<span>{{.Label}}</span>	
			<input type="string" gl-input="Value" />
		</div>
	`
}
