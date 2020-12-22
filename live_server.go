package golive

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
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
	s.Wire.ActivateLivePage(session, p)

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

	errors := make(chan error)
	exit := make(chan int)

	exited := false

	outMessages, err := s.Wire.GetOutChannel(sKey)

	if err != nil {
		//errors <- err
		panic(err)
	}

	go func() {
		for {
			select {
			case msg := <-*outMessages:
				if msg.Type == "CLOSE" {
					exit <- 1
				}
				err := c.WriteJSON(msg)
				if err != nil {
					errors <- err
				}

			case <-exit:
				exited = true
				err := c.Close()

				if err != nil {
					errors <- err
				}

				return
			case err := <-errors:
				fmt.Println("level=error", err)
				break

			default:
				break
			}
		}
	}()

	c.SetCloseHandler(func(code int, text string) error {
		exit <- 1
		return nil
	})

	for {
		if exited {
			return
		}

		inMsg := InMessage{}
		err := c.ReadJSON(&inMsg)

		if err != nil {
			errors <- err
		}

		err = s.Wire.IngestMessage(sKey, inMsg)
		if err != nil {
			errors <- err
		}
	}
}
