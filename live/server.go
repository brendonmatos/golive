package live

import (
	"context"
	"fmt"
	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/differ"
	"github.com/brendonmatos/golive/live/component"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	// Wire ...
	Wire *Wire

	// CookieName ...
	CookieName string
	Log        golive.Log
}

type Response struct {
	Rendered string
	Session  string
}

func NewServer() *Server {
	logger := golive.NewLoggerBasic()
	return &Server{
		Wire:       NewWire(),
		CookieName: "_csrf_token",
		Log:        logger.Log,
	}
}

func (s *Server) HandleFirstRequest(lc *component.Component, c PageContent) (*Response, error) {

	/* Create session to the new user */
	sessionKey, session, err := s.Wire.CreateSession()
	if err != nil {
		return nil, err
	}

	s.Log(golive.LogInfo, "http request", golive.LogEx{"Component": lc.Name, "session": sessionKey})

	session.log = s.Log

	// Instantiate a page to attach to a session
	p := NewLivePage(lc)

	// Set page content
	p.SetContent(c)

	// activation should be before mount,
	// because in activation will setup page channels
	// that will be needed in mount
	session.ActivatePage(p)

	p.Create()

	rendered, err := p.Render()

	if err != nil {
		return &Response{
			Rendered: "<h1> Page with error </h1>",
			Session:  "",
		}, fmt.Errorf("page render: %w", err)
	}

	return &Response{Rendered: rendered, Session: sessionKey}, nil
}

func (s *Server) HandleHTMLRequest(ctx *fiber.Ctx, lc *component.Component, c PageContent) {

	lr, err := s.HandleFirstRequest(lc, c)

	if lr == nil {
		s.Log(golive.LogPanic, "no live page", golive.LogEx{"error": err})
		return
	}

	if err != nil {
		s.Log(golive.LogError, "handle html request", golive.LogEx{"error": err})
		ctx.Response().SetStatusCode(500)
		return
	}

	ctx.Cookie(&fiber.Cookie{
		Name:    s.CookieName,
		Value:   lr.Session,
		Expires: time.Now().Add(24 * time.Hour),
	})

	ctx.Response().Header.SetContentType("text/html")
	ctx.Response().AppendBodyString(lr.Rendered)
}

func (s *Server) CreateHTMLHandler(f func() *component.Component, c PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		lc := f()
		lc.Log = s.Log

		s.HandleHTMLRequest(ctx, lc, c)
		return nil
	}
}

// HTTPMiddleware Middleware to run on HTTP requests.
type HTTPMiddleware func(next HTTPHandlerCtx) HTTPHandlerCtx

// HTTPHandlerCtx HTTP Handler with a page level context.
type HTTPHandlerCtx func(ctx *fiber.Ctx, pageCtx context.Context)

func (s *Server) CreateHTMLHandlerWithMiddleware(f func(ctx context.Context) *component.Component, content PageContent,
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
		lc.Log = s.Log

		s.HandleHTMLRequest(c, lc, content)

		return nil
	}
}

func (s *Server) HandleWSRequest(c *websocket.Conn) {
	defer func() {
		payload := recover()
		if payload != nil {
			s.Log(golive.LogWarn, fmt.Sprintf("ws request panic recovered: %v", payload), nil)
		}
	}()

	c.EnableWriteCompression(true)

	sessionKey := c.Cookies(s.CookieName)

	s.Log(golive.LogInfo, "websocket open", golive.LogEx{"session": sessionKey})

	session := s.Wire.GetSession(sessionKey)

	if session == nil || session.Status != SessionNew {
		s.Log(golive.LogWarn, "session not found", golive.LogEx{"session": sessionKey})

		var msg differ.PatchBrowser
		msg.Type = EventLiveError
		msg.Message = ErrorSessionNotFound
		if err := c.WriteJSON(msg); err != nil {
			s.Log(golive.LogError, "handle ws request: write json", golive.LogEx{"error": err})
		}

		if err := c.Close(); err != nil {
			s.Log(golive.LogError, "close websocket connection", golive.LogEx{"error": err})
		}

		s.Log(golive.LogInfo, "websocket close", golive.LogEx{"session": sessionKey})

		return
	}

	session.Status = SessionOpen
	exit := make(chan int)
	exited := false

	go func() {
		for {
			select {
			case msg := <-session.OutChannel:
				s.Log(golive.LogDebug, "message out", golive.LogEx{"msg": msg, "session": sessionKey})

				if err := c.WriteJSON(msg); err != nil {
					s.Log(golive.LogError, "handle ws request: write json", golive.LogEx{"error": err})
				}
			case <-exit:
				exited = true

				session.Status = SessionClosed

				if err := c.Close(); err != nil {
					s.Log(golive.LogError, "close websocket connection", golive.LogEx{"error": err})
				}

				if err := session.LivePage.EntryComponent.Unmount(); err != nil {
					s.Log(golive.LogError, "handle ws request: kill page", golive.LogEx{"error": err})
				}

				s.Wire.DeleteSession(sessionKey)

				s.Log(golive.LogInfo, "websocket close", golive.LogEx{"session": sessionKey})

				return
			}
		}
	}()

	c.SetCloseHandler(func(code int, text string) error {
		// Close codes defined in RFC 6455, section 11.7.
		s.Log(golive.LogTrace, "ws close handler", golive.LogEx{"code": code, "text": text})

		exit <- 1
		return nil
	})

	for {
		if exited {
			return
		}

		inMsg := BrowserEvent{}

		// Loop blocks here
		if err := c.ReadJSON(&inMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				// This seems to happen when running in Docker
				if !exited {
					s.Log(golive.LogWarn, "handle ws request: unexpected connection close", nil)

					exit <- 1
				}

				return
			}

			s.Log(golive.LogError, "handle ws request: read json", golive.LogEx{"error": err})

			continue
		}

		s.Log(golive.LogDebug, "message in", golive.LogEx{"msg": inMsg, "session": sessionKey})

		if err := session.IngestMessage(inMsg); err != nil {
			s.Log(golive.LogError, "handle ws request: ingest message ", golive.LogEx{"error": err})
		}
	}
}
