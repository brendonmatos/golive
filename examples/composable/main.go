package main

import (
	"fmt"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateStaticPageRender(func() *component.Component {
		counter := NewCounter(1)
		return counter
	}, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWebSocketConnection))

	err := app.Listen(":3000")

	if err != nil {
		fmt.Println(err)
	}
}
