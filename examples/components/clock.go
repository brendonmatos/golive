package components

import (
	"github.com/brendonmatos/golive/live"
	"time"
)

type Clock struct {
	live.Wrapper
	ActualTime string
}

func NewClock() *live.Component {
	return live.NewLiveComponent("Clock", &Clock{
		ActualTime: formattedActualTime(),
	})
}

func formattedActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *live.Component) {
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

func (c *Clock) TemplateHandler(_ *live.Component) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
