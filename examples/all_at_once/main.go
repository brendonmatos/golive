package main

import (
	"github.com/brendonmatos/golive/examples/components"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/renderer"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Home struct {
}

func NewHome() *live.Component {
	c := live.DefineComponent("home")

	c.UseComponent("Clock", components.NewClock)
	c.UseComponent("Todo", components.NewTodo)
	c.UseComponent("Slider", components.NewSlider)

	c.SetState(&Home{})

	c.UseRender(renderer.NewTemplateRenderer(`
		<div>
			{{ render "Clock"  }}
			{{ render "Todo" }}
			{{ render "Slider" }}
		</div>
	`))

	return c
}

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(NewHome, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	_ = app.Listen(":3000")

}
