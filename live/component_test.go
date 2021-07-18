package live

import (
	"github.com/brendonmatos/golive/live/renderer"
	"time"
)

type Clock struct {
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func NewClock() *Component {
	c := DefineComponent("Clock")

	c.SetState(&Clock{})

	OnMounted(c, func() {
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

type TestComp struct {
	Wrapper
}

func (tc *TestComp) TemplateHandler(_ *Component) string {
	return `
		<div>
			<div></div>
			<div>
				<div></div>
			</div>
			<div></div>
			<div></div>
		</div>
	`
}
