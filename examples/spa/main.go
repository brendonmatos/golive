package main

import (
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/live"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()
	liveServer := live.NewServer()
	basicLogger := golive.NewLoggerBasic()
	basicLogger.Level = golive.LogDebug
	liveServer.Log = basicLogger.Log
	app.Get("/ws", websocket.New(liveServer.HandleWebSocketConnection))
	app.Get("/*", liveServer.CreateStaticPageRender(NewRouter, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))
	_ = app.Listen(":3000")
}
