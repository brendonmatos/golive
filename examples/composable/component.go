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

	c := live.DefineComponent("composed")

	counter := &Counter{
		Actual: actual,
	}

	c.SetState(counter)

	live.OnMounted(c, func() {
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

	err := c.UseRender(renderer.NewTemplateRenderer(`
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
