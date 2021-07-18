package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/renderer"
	"time"
)

type Clock struct {
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func NewClock() *live.Component {
	c := live.DefineComponent("Clock")

	c.SetState(&Clock{})

	live.OnMounted(c, func() {
		go func() {
			for {
				if c.Context.Closed {
					return
				}
				time.Sleep(time.Second)
				c.Update()
			}
		}()
	})

	err := c.UseRender(renderer.NewTemplateRenderer(`
			<div>
				<span>Time: {{ .ActualTime }}</span>
			</div>
		`))
	if err != nil {
		return nil
	}

	return c
}
