package live

import (
	"fmt"
	"time"

	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/util"
	"github.com/brendonmatos/golive/live/wire"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	Sessions map[string]*Session

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
		Sessions:   map[string]*Session{},
		CookieName: "_csrf_token",
		Log:        logger.Log,
	}
}

func (s *Server) CreateSession() (string, *Session) {
	key, _ := util.GenerateRandomString(48)
	ns := NewSession()
	s.Sessions[key] = ns
	return key, ns
}

func (s *Server) GetSession(key string) *Session {
	return s.Sessions[key]
}

func (s *Server) DeleteSession(key string) {
	ss := s.GetSession(key)

	if err := ss.Wire.End(); err != nil {
		s.Log(golive.LogError, "handle ws request: kill page", golive.LogEx{"error": err})
	}

	delete(s.Sessions, key)
}

func (s *Server) HandleFirstRequest(lc *component.Component, c PageContent) (*Response, error) {

	/* Create session to the new user */
	sessionKey, session := s.CreateSession()

	s.Log(golive.LogInfo, "http request", golive.LogEx{"Component": lc.Name, "session": sessionKey})

	session.log = s.Log

	// Instantiate a page to attach to a session
	p := NewLivePage(lc)

	// Set page content
	p.SetContent(c)

	// activation should be before mount,
	// because in activation will setup page channels
	// that will be needed in mount
	err := session.WireComponent(lc)
	if err != nil {
		return nil, fmt.Errorf("session wire component: %w", err)
	}

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

func (s *Server) CreateStaticPageRender(f func() *component.Component, c PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		lc := f()
		lc.Log = s.Log

		component.Provide(lc, "fiber_ctx", ctx)

		s.HandleHTMLRequest(ctx, lc, c)
		return nil
	}
}

func (s *Server) HandleWebSocketConnection(c *websocket.Conn) {
	defer func() {
		payload := recover()
		if payload != nil {
			s.Log(golive.LogWarn, fmt.Sprintf("ws request panic recovered: %v", payload), nil)
		}
	}()
	c.EnableWriteCompression(true)
	sessionKey := c.Cookies(s.CookieName)
	s.Log(golive.LogInfo, "websocket open", golive.LogEx{"session": sessionKey})
	session := s.GetSession(sessionKey)

	if session == nil || session.Status != SessionNew {
		s.Log(golive.LogWarn, "session not found", golive.LogEx{"session": sessionKey})
		var msg wire.ToBrowser
		msg.Type = wire.ToBrowserLiveError
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

	session.SetOpen()
	exit := make(chan int)
	exited := false

	w := session.Wire

	go func() {
		for {
			select {
			case msg := <-w.ToBrowser:
				s.Log(golive.LogDebug, "message out", golive.LogEx{"msg": msg, "session": sessionKey})
				if err := c.WriteJSON(msg); err != nil {
					s.Log(golive.LogError, "handle ws request: write json", golive.LogEx{"error": err})
				}
			case <-exit:
				exited = true
				session.SetClosed()
				if err := c.Close(); err != nil {
					s.Log(golive.LogError, "close websocket connection", golive.LogEx{"error": err})
				}
				s.DeleteSession(sessionKey)
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

		inMsg := wire.FromBrowser{}

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

		w.HandleFromBrowser(&inMsg)
	}
}
