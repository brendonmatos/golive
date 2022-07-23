package main

import (
	"github.com/brendonmatos/golive/examples/components"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Home struct {
}

func NewHome() *component.Component {
	c := component.DefineComponent("home", func(ctx *component.Context) renderer.Renderer {

		return renderer2.NewTemplateRenderer(`
			<div>
				{{ render "Clock"  }}
				{{ render "Todo" }}
				{{ render "Slider" }}
			</div>
		`)
	})

	c.UseComponent("Clock", components.NewClock)
	c.UseComponent("Todo", components.NewTodo)
	c.UseComponent("Slider", components.NewSlider)

	c.SetState(&Home{})

	err := c.UseRender()
	if err != nil {
		return nil
	}

	return c
}

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateStaticPageRender(NewHome, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWebSocketConnection))

	_ = app.Listen(":3000")

}
