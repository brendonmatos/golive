package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/renderer"
)

type Counter struct {
	Actual int
}

func (c *Counter) Increase() {
	c.Actual++
}

func NewCounter(actual int) *live.Component {

	c := live.DefineComponent("composed")

	counter := &Counter{
		Actual: actual,
	}

	c.SetState(counter)

	err := c.UseRender(renderer.NewRenderer(renderer.NewTemplateRenderer(`
		<div>
			<button gl-click="Increase">Increase</button>
			<div>{{ .Actual }}</div>
			<input type="text" gl-input="Actual" />
		</div>
	`)))

	if err != nil {
		panic(err)
	}

	return c
}