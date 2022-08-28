package main

import (
	"fmt"
	"time"

	"github.com/brendonmatos/golive/impl/live_fiber"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	liveFiber := live_fiber.NewFiberServer()

	app.Get("/", liveFiber.CreateStaticPageRender(NewClock, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveFiber.HandleWebSocketConnection))

	_ = app.Listen(":3000")

}

type Clock struct {
	Message string
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func NewClock() *component.Component {
	return component.DefineComponent("Clock", func(ctx *component.Context) string {

		state := component.UseState(ctx, &Clock{
			Message: "Hello World",
		})

		update := component.UseUpdate(ctx)

		component.UseOnMounted(ctx, func() {
			go func() {
				for {
					time.Sleep(time.Second)
					if ctx.Closed {
						return
					}
					update()
				}
			}()
		})

		return fmt.Sprintf(`<div>
			<span>Time: %s</span>
			<input type="text" value="%s" gl-input="Message" />
			<div>-->%s<--</div>
		</div>`, state.ActualTime(), state.Message, state.Message)
	})
}
