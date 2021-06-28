package components

import (
	"github.com/brendonmatos/golive/live"
)

type Slider struct {
	live.Wrapper
	Size float32
}

func NewSlider() *live.Component {
	return live.NewLiveComponent("Slider", &Slider{
		Size: 40,
	})
}

func (t *Slider) Size2() float32 {
	return t.Size * 2
}

func (t *Slider) Size3() float32 {
	return t.Size * t.Size * 0.3
}

func (t *Slider) TemplateHandler(_ *live.Component) string {
	return `
		<div>
			<input gl-input="Size" type="range" value="{{.Size}}"/>
			<div class="" style="background-color: black; width: {{ .Size3 }}px; height: {{ .Size2 }}px">
				<div style="background-color: red; width: {{ .Size2 }}px; height: {{.Size2}}px" >
				</div>
			</div>
		</div>	
	`
}
