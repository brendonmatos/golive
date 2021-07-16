package main

import (
	"fmt"
	"github.com/brendonmatos/golive/live"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	counter := NewCounter(1)

	app.Get("/", liveServer.CreateHTMLHandler(func() *live.Component {
		return counter
	}, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	err := app.Listen(":3000")

	if err != nil {
		fmt.Println(err)
	}
}
