package golive

import (
	"github.com/gofiber/fiber/v2"
	//"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"time"
)

type LiveServer struct {
	// Wire ...
	Wire *LiveWire

	// CookieName ...
	CookieName string
}

type LiveResponse struct {
	Rendered string
	Session  SessionKey
}

func NewServer() *LiveServer {
	return &LiveServer{
		Wire:       NewWire(),
		CookieName: "_csrf_token",
	}
}

func (s *LiveServer) HandleFirstRequest(lc *LiveComponent, c PageContent) LiveResponse {

	/* Create session to the new user */
	session, _ := s.Wire.CreateSession()

	/* Instantiate a page to attach to a session */
	p := NewLivePageToComponent(session, lc)

	/* Mounts the page */
	p.Mount()

	/*  */
	rendered := p.FirstRender(c)
	s.Wire.SetSession(session, p)
	return LiveResponse{Rendered: rendered, Session: session}
}

func (s *LiveServer) HandleHTMLRequest(ctx *fiber.Ctx, lc *LiveComponent, c PageContent) {

	lr := s.HandleFirstRequest(lc, c)

	ctx.Cookie(&fiber.Cookie{
		Name:    s.CookieName,
		Value:   string(lr.Session),
		Expires: time.Now().Add(24 * time.Hour),
	})
	ctx.Response().Header.SetContentType("text/html")
	ctx.Response().AppendBodyString(lr.Rendered)

	return
}

func (s *LiveServer) CreateHTMLHandler(f func() *LiveComponent, c PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		s.HandleHTMLRequest(ctx, f(), c)
		return nil
	}
}

func (s *LiveServer) HandleWSRequest(c *websocket.Conn) {

	c.EnableWriteCompression(true)

	cookie := c.Cookies(s.CookieName)
	sKey := SessionKey(cookie)

	quit := make(chan int)

	go func() {
		for {
			select {
			case <-quit:
				close(quit)
				return
			case msg := <-s.Wire.OutChannel:
				if msg.OutMessage.Type == "CLOSE" {
					quit <- 1
				}
				err := c.WriteJSON(msg.OutMessage)
				if err != nil {
					quit <- 1
				}
			}
		}
	}()

	for {
		inMsg := InMessage{}
		err := c.ReadJSON(&inMsg)

		if err != nil {
			log.Printf("read msg error: %v", err)
			quit <- 1
			break
		}

		err = s.Wire.HandleMessage(sKey, inMsg)
		if err != nil {
			quit <- 1
		}
	}
}
