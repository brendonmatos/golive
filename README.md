# GoLive

## 💻 Reactive HTML Server Side Rendered by GoLang over WebSockets 🚀

Use just GoLang to program reactive front-ends!

**Project WIP. Your feedback and PR are welcome!**

![](examples/slider/slider.gif)

## How?

1. Render Server Side HTML
2. Connect to same server using Websocket
3. Send user events
4. Change state of [component](component.go) in server
5. Render Component and get [diff](diff.go)
6. Update instructions are sent to the browser

## What is this lib suitable for?

This project it's strongly inspired by Elixir Phoenix LiveView. But, to be fair, i don't see being used for usual front-end applications replacing Vue or React. Otherwise, just a reactive view to build an internal tool or ease the management/config/dashboard from **UI**, can be the perfect use case for this library.
Just to exemplify an use case: [Gorse Project](https://github.com/zhenghaoz/gorse) which has a dashboard user interface, could have been built using GoLive which would make all the stats live with minimal effort, all integrating with the real data. No need to create API for [Front-end](https://github.com/gorse-io/dashboard).

## Getting Started

- [Extended Version Todo Example](https://github.com/SamHennessy/golive-example)
- [Project Examples](https://github.com/brendonmatos/golive/tree/master/examples)
- [GoBook - Interactive Go REPL in browser](https://github.com/brendonmatos/gobook)

## Component Example

```go
package components

import (
	"time"

	c "github.com/brendonmatos/golive/live/component"
	r "github.com/brendonmatos/golive/live/component/renderer"
)

type Clock struct {
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func NewClock() *c.Component {
	clock := c.NewLiveComponent("Clock", &Clock{})

	c.OnMounted(clock, func(_ *c.Context) {
		go func() {
			for {
				if clock.Context.Closed {
					return
				}
				time.Sleep(time.Second)
				clock.Update()
			}
		}()
	})

	clock.UseRender(r.NewTemplateRenderer(`
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`))
	return clock
}

```

### Server Example

```go

package main

import (
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/examples/components"
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

### That's it!

![](examples/clock/demo.gif)

## More Examples

### Slider

![](examples/slider/slider.gif)

### Simple todo

![](examples/todo/todo.gif)

### All at once using components!

![](examples/all_at_once/all_at_once.gif)

### GoBook

![](examples/gobook.gif)

[Go to repo](https://github.com/brendonmatos/gobook)
