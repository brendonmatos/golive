package components

import (
	"time"

	"github.com/brendonmatos/golive"
)

type Clock struct {
	golive.LiveComponentWrapper
	ActualTime string
}

func NewClock() *golive.LiveComponent {
	return golive.NewLiveComponent("Clock", &Clock{
		ActualTime: formattedActualTime(),
	})
}

func formattedActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *golive.LiveComponent) {
	go func() {
		for {
			if l.Exited {
				return
			}
			c.ActualTime = formattedActualTime()
			time.Sleep(time.Second)
			c.Commit()
		}
	}()
}

func (c *Clock) TemplateHandler(_ *golive.LiveComponent) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
