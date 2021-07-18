package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/renderer"
	"time"
)

type Counter struct {
	Actual int
}

func (c *Counter) Increase() {
	c.Actual++
}

func NewCounter(actual int) *live.Component {

	component := live.DefineComponent("composed")

	counter := &Counter{
		Actual: actual,
	}

	component.SetState(counter)

	live.OnMounted(component, func() {
		go func() {
			for {
				counter.Actual = counter.Actual + 3000
				time.Sleep(time.Millisecond)
				component.Update()
			}
		}()
	})

	err := component.UseRender(renderer.NewTemplateRenderer(`
		<div>
			<button gl-click="Increase">Increase</button>
			<div>{{ .Actual }}</div>
			<input type="text" gl-input="Actual" />
		</div>
	`))

	if err != nil {
		panic(err)
	}

	return component
}
