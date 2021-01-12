package components

import "github.com/brendonmatos/golive"

type DynamicInput struct {
	golive.LiveComponentWrapper
	Label string
}

func NewDynamicInput() *golive.LiveComponent {
	return golive.NewLiveComponent("DynamicInput", &DynamicInput{
		Label: "",
	})
}

func (d *DynamicInput) TemplateHandler(_ *golive.LiveComponent) string {
	return `
		<div>
			<input type="string" go-live-input="Label" />
			<span>{{.Label}}</span>	
		</div>
	`
}
