package components

import (
	"time"

	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
)

type Clock struct {
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func NewClock() *component.Component {
	c := component.DefineComponent("Clock")

	c.SetState(&Clock{})

	component.OnMounted(c, func(_ *component.Context) {
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
