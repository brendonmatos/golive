package components

import "github.com/brendonmatos/golive"

type DynamicInputProps struct {
	Value *string
	Label string
}

type DynamicInput struct {
	golive.LiveComponentWrapper
	Value *string
	Label string
}

func NewDynamicInput(props DynamicInputProps) *golive.LiveComponent {
	return golive.NewLiveComponent("DynamicInput", &DynamicInput{
		Value: props.Value,
		Label: props.Label,
	})
}

func (d *DynamicInput) TemplateHandler(_ *golive.LiveComponent) string {
	return `
		<div>
			<span>{{.Label}}</span>	
			<input type="string" go-live-input="Value" />
		</div>
	`
}
