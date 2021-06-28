package components

import (
	"github.com/brendonmatos/golive/live"
)

type DynamicInputProps struct {
	Value *string
	Label string
}

type DynamicInput struct {
	live.Wrapper
	Value *string
	Label string
}

func NewDynamicInput(props DynamicInputProps) *live.Component {
	return live.NewLiveComponent("DynamicInput", &DynamicInput{
		Value: props.Value,
		Label: props.Label,
	})
}

func (d *DynamicInput) TemplateHandler(_ *live.Component) string {
	return `
		<div>
			<span>{{.Label}}</span>	
			<input type="string" gl-input="Value" />
		</div>
	`
}
