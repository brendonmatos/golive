package main

import (
	"fmt"
	components "github.com/brendonmatos/golive/examples/components"
	"github.com/brendonmatos/golive/live"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Home struct {
	live.Wrapper
	Clock  *live.Component
	Todo   *live.Component
	Slider *live.Component
}

func NewHome() *live.Component {
	return live.NewLiveComponent("Home", &Home{
		Clock:  components.NewClock(),
		Todo:   components.NewTodo(),
		Slider: components.NewSlider(),
	})
}

func (h *Home) Mounted(_ *live.Component) {
	return
}

func (h *Home) TemplateHandler(_ *live.Component) string {
	return `
	<div>
		{{render .Clock}}
		{{render .Todo}}
		{{render .Slider}}
	</div>
	`
}

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(NewHome, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	fmt.Println(app.Listen(":3000"))

}
