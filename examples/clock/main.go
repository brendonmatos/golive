package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(NewClock, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	_ = app.Listen(":3000")

}
