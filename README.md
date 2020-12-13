# GoLive

![](demo.gif)

**Any suggestions are absolutely welcome**

This project it's strongly inspired by Elixir Phoenix Live View. I'm writing this as a Side Project to solve some issues that I'm having in production in my full-time Job.

## Component Example
```go

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
```

### Server Example
```go
  
package main

import (
	"github.com/brendonferreira/golive"
	"github.com/brendonferreira/golive/examples/components"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()
	liveServer := golive.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(components.NewClock, golive.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	_ = app.Listen(":3000")
}
```

**That it!**
