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
	Log        Log
}

type LiveResponse struct {
	Rendered string
	Session  string
}

func NewServer() *LiveServer {
	return &LiveServer{
		Wire:       NewWire(),
		CookieName: "_csrf_token",
		Log:        NewLoggerBasic().Log,
	}
}

func (s *LiveServer) HandleFirstRequest(lc *LiveComponent, c PageContent) (*LiveResponse, error) {
	/* Create session to the new user */
	sessionKey, session, err := s.Wire.CreateSession()
	if err != nil {
		return nil, err
	}

	session.log = s.Log

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
		s.Log(LogPanic, "no live page", logEx{"error": err})

		return
	}

	s.Log(LogInfo, "http request", logEx{"component": lc.Name, "session": lr.Session})

	ctx.Cookie(&fiber.Cookie{
		Name:    s.CookieName,
		Value:   lr.Session,
		Expires: time.Now().Add(24 * time.Hour),
	})
	ctx.Response().Header.SetContentType("text/html")
	ctx.Response().AppendBodyString(lr.Rendered)

	if err != nil {
		s.Log(LogError, "handle html request", logEx{"error": err})
		ctx.Response().SetStatusCode(500)
	}
}

func (s *LiveServer) CreateHTMLHandler(f func() *LiveComponent, c PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		lc := f()
		lc.log = s.Log

		s.HandleHTMLRequest(ctx, lc, c)
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

		lc := f(ctx)
		lc.log = s.Log

		s.HandleHTMLRequest(c, lc, content)

		return nil
	}
}

func (s *LiveServer) HandleWSRequest(c *websocket.Conn) {
	c.EnableWriteCompression(true)

	sessionKey := c.Cookies(s.CookieName)
	session := s.Wire.GetSession(sessionKey)

	s.Log(LogInfo, "websocket open", logEx{"session": sessionKey})

	errors := make(chan error)
	exit := make(chan int)

	exited := false

	go func() {
		for {
			select {
			case msg := <-session.OutChannel:
				s.Log(LogDebug, "message out", logEx{"msg": msg, "session": sessionKey})

				err := c.WriteJSON(msg)
				if err != nil {
					errors <- err
				}

			case <-exit:
				exited = true

				if err := c.Close(); err != nil {
					errors <- fmt.Errorf("close websocket connection: %w", err)
				}

				if err := session.LivePage.entry.Kill(); err != nil {
					errors <- fmt.Errorf("kill page entry component: %w", err)
				}

				s.Log(LogInfo, "websocket close", logEx{"session": sessionKey})

				return
			case err := <-errors:
				s.Log(LogError, "websocket error", logEx{"error": err, "session": sessionKey})
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

		s.Log(LogDebug, "message in", logEx{"msg": inMsg, "session": sessionKey})

		err = session.IngestMessage(inMsg)
		if err != nil {
			errors <- err
		}
	}
}
