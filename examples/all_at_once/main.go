package main

import (
	"fmt"

	"github.com/brendonmatos/golive"
	components "github.com/brendonmatos/golive/examples/components"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Home struct {
	golive.LiveComponentWrapper
	Clock  *golive.LiveComponent
	Todo   *golive.LiveComponent
	Slider *golive.LiveComponent
}

func NewHome() *golive.LiveComponent {
	return golive.NewLiveComponent("Home", &Home{
		Clock:  components.NewClock(),
		Todo:   components.NewTodo(),
		Slider: components.NewSlider(),
	})
}

func (h *Home) Mounted(_ *golive.LiveComponent) {
	return
}

func (h *Home) TemplateHandler(_ *golive.LiveComponent) string {
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
	liveServer := golive.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(NewHome, golive.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	fmt.Println(app.Listen(":3000"))

}
