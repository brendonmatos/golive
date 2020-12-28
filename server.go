package golive

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type LiveServer struct {
	// Wire ...
	Wire *LiveWire

	// CookieName ...
	CookieName string
}

type LiveResponse struct {
	Rendered string
	Session  string
}

func NewServer() *LiveServer {
	return &LiveServer{
		Wire:       NewWire(),
		CookieName: "_csrf_token",
	}
}

func (s *LiveServer) HandleFirstRequest(lc *LiveComponent, c PageContent) (*LiveResponse, error) {

	/* Create session to the new user */
	sessionKey, session, err := s.Wire.CreateSession()

	if err != nil {
		return nil, err
	}

	/* Instantiate a page to attach to a session */
	p := NewLivePage(lc)
	p.SetContent(c)

	// 1.
	p.Prepare()

	// 2.
	p.Mount()

	/*  */
	rendered, err := p.Render()

	if err != nil {
		return &LiveResponse{
			Rendered: "<h1> Page with error </h1>",
			Session:  "",
		}, err
	}

	/*  */
	session.ActivatePage(p)

	return &LiveResponse{Rendered: rendered, Session: sessionKey}, nil
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

// HTTPMiddleware Middleware to run on HTTP requests.
type HTTPMiddleware func(next HTTPHandlerCtx) HTTPHandlerCtx

// HTTPHandlerCtx HTTP Handler with a page level context.
type HTTPHandlerCtx func(ctx *fiber.Ctx, pageCtx context.Context)

func (s *LiveServer) CreateHTMLHandlerWithMiddleware(f func(ctx context.Context) *LiveComponent, content PageContent,
	middlewares ...HTTPMiddleware) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		// TODO: move chain building out so it only happens once - Sam H.
		if len(middlewares) != 0 {
			// Reassign the context back to this scope to capture changes to it
			end := func(_ *fiber.Ctx, pCtx context.Context) {
				ctx = pCtx
			}
			h := middlewares[len(middlewares)-1](end)

			// build chain
			for i := len(middlewares) - 2; i >= 0; i-- {
				h = middlewares[i](h)
			}
			// trigger chain
			h(c, ctx)
		}

		s.HandleHTMLRequest(c, f(ctx), content)

		return nil
	}
}

func (s *LiveServer) HandleWSRequest(c *websocket.Conn) {

	c.EnableWriteCompression(true)

	sessionKey := c.Cookies(s.CookieName)
	session := s.Wire.GetSession(sessionKey)

	errors := make(chan error)
	exit := make(chan int)

	exited := false

	go func() {
		for {
			select {
			case msg := <-session.OutChannel:
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

		err = session.IngestMessage(inMsg)
		if err != nil {
			errors <- err
		}
	}
}
