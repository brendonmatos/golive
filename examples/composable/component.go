package main

import (
	"github.com/brendonmatos/golive/live/component"
	renderer2 "github.com/brendonmatos/golive/live/component/renderer"
	"time"
)

type Counter struct {
	Actual int
}

func (c *Counter) Increase() {
	c.Actual++
}

func NewCounter(actual int) *component.Component {

	c := component.DefineComponent("composed")

	counter := &Counter{
		Actual: actual,
	}

	c.SetState(counter)

	component.OnMounted(c, func() {
		go func() {
			for {
				if c.Context.Closed {
					return
				}
				counter.Actual = counter.Actual + 3000
				time.Sleep(time.Second / 60)
				c.Update()
			}
		}()
	})

	err := c.UseRender(renderer2.NewTemplateRenderer(`
		<div>
			<button gl-click="Increase">Increase</button>
			<div>{{ .Actual }}</div>
			<input type="text" gl-input="Actual" />
		</div>
	`))

	if err != nil {
		panic(err)
	}

	return c
}
