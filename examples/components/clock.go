package components

import (
	"github.com/brendonferreira/golive"
	"time"
)

type Clock struct {
	golive.LiveComponentWrapper
	ActualTime string
}

func NewClock() *golive.LiveComponent {
	return golive.NewLiveComponent("Clock", &Clock{
		ActualTime: "91230192301390193",
	})
}

func (t *Clock) Mounted(_ *golive.LiveComponent) {
	t.Tick()
}

func (t *Clock) Tick() {

	go func() {
		t.ActualTime = time.Now().Format(time.RFC3339Nano)
		t.Commit()

		time.Sleep((time.Second * 1) / 60)
		go t.Tick()
	}()
}

func (t *Clock) TemplateHandler() string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
