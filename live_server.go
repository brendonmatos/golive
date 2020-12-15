package golive

import (
	"github.com/gofiber/fiber/v2"
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

func (s *LiveServer) HandleFirstRequest(lc *LiveComponent, c PageContent) (*LiveResponse, error) {

	/* Create session to the new user */
	session, err := s.Wire.CreateSession()

	if err != nil {
		return nil, err
	}

	/* Instantiate a page to attach to a session */
	p := NewLivePage(session, lc)

	p.Prepare()
	p.Mount()

	/*  */
	rendered, err := p.FirstRender(c)

	if err != nil {
		return &LiveResponse{
			Rendered: "<h1> Page with error </h1>",
			Session:  "",
		}, err
	}

	/*  */
	s.Wire.SetSession(session, p)

	return &LiveResponse{Rendered: rendered, Session: session}, nil
}

func (s *LiveServer) HandleHTMLRequest(ctx *fiber.Ctx, lc *LiveComponent, c PageContent) {

	lr, err := s.HandleFirstRequest(lc, c)

	if lr == nil {
		panic(err)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:    s.CookieName,
		Value:   string(lr.Session),
		Expires: time.Now().Add(24 * time.Hour),
	})
	ctx.Response().Header.SetContentType("text/html")
	ctx.Response().AppendBodyString(lr.Rendered)

	if err != nil {
		ctx.Response().SetStatusCode(500)
	}

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
			panic(err)
		}
	}
}
