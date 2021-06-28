package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/renderer"
	"github.com/brendonmatos/golive/live/state"
)

type Counter struct {
	actual int
}

func (c *Counter) Increase() {
	c.actual++
}

func NewCounter(actual int) *live.Component {

	c := live.DefineComponent("composed")

	counter := &Counter{
		actual: actual,
	}

	c.SetState(counter)

	r := renderer.NewStaticRenderer(`
			<div>
				<button gl-click="Increase">Increase</button>
				<div>%d</div>
			</div>
		`,
		func(s *state.State) []interface{} {
			return []interface{}{counter.actual}
		})

	c.Renderer = renderer.NewRenderer(c.Name, r)

	return c
}
