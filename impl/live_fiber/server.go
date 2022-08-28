package live_fiber

import (
	"time"

	"github.com/brendonmatos/golive"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/util"
	"github.com/brendonmatos/golive/live/wire"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type FiberServer struct {
	Server *live.Server
	// CookieName ...
	Sessions   map[string]*live.Session
	CookieName string
	Log        golive.Log
}

type Response struct {
	Rendered string
	Session  string
}

func NewFiberServer() *FiberServer {
	logger := golive.NewLoggerBasic()

	return &FiberServer{
		Server:     live.NewServer(),
		Sessions:   map[string]*live.Session{},
		CookieName: "_csrf_token",
		Log:        logger.Log,
	}
}

func (s *FiberServer) GetSession(key string) *live.Session {
	return s.Sessions[key]
}

func (s *FiberServer) StoreSession(ls *live.Session) string {
	key, _ := util.GenerateRandomString(48)
	s.Sessions[key] = ls
	return key
}

func (s *FiberServer) DeleteSession(key string) {
	delete(s.Sessions, key)
}

func (s *FiberServer) CreateLiveComponent(a func(ctx *component.Context) string, c live.PageContent) func(ctx *fiber.Ctx) error {
	return s.CreateStaticPageRender(func() *component.Component {
		return component.DefineComponent("Clock", func(ctx2 *component.Context, props *interface{}) string {
			return a(ctx2)
		})
	}, c)
}

func (s *FiberServer) CreateStaticPageRender(f func() *component.Component, c live.PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		lc := f()
		lc.Log = s.Log

		component.Provide(lc, "fiber_ctx", ctx)
		ctx.Response().Header.SetContentType("text/html")

		session := s.Server.CreateSession()
		sessionKey := s.StoreSession(session)

		rendered, err := s.Server.CreateLivePage(session, lc, c)

		if err != nil {
			ctx.Response().AppendBodyString("<h1> Page with error </h1>")
			ctx.Response().AppendBodyString("<strong style='color: red'>" + err.Error() + "</strong>")
			return err
		}

		ctx.ClearCookie(s.CookieName)

		ctx.Cookie(&fiber.Cookie{
			Name:    s.CookieName,
			Value:   sessionKey,
			Expires: time.Now().Add(24 * time.Hour),
		})

		ctx.Response().AppendBodyString(*rendered)

		return nil
	}
}

func (s *FiberServer) HandleWebSocketConnection(c *websocket.Conn) {
	c.EnableWriteCompression(true)
	sessionKey := c.Cookies(s.CookieName)
	s.Log(golive.LogInfo, "websocket open", golive.LogEx{"session": sessionKey})
	session := s.GetSession(sessionKey)

	if session == nil {
		s.Log(golive.LogWarn, "session not found", golive.LogEx{"session": sessionKey})
		msg := make(map[string]string)
		msg["Type"] = string(wire.FromServerLiveError)
		msg["Message"] = "Session not found"
		if err := c.WriteJSON(msg); err != nil {
			s.Log(golive.LogError, "handle ws request: write json", golive.LogEx{"error": err})
		}
		if err := c.Close(); err != nil {
			s.Log(golive.LogError, "close websocket connection", golive.LogEx{"error": err})
		}
		s.Log(golive.LogInfo, "websocket close", golive.LogEx{"session": sessionKey})
		return
	}

	w := session.Wire

	c.SetCloseHandler(func(code int, text string) error {
		// Close codes defined in RFC 6455, section 11.7.
		s.Log(golive.LogInfo, "ws close handler", golive.LogEx{"session": sessionKey})
		session.Close()
		return nil
	})

	go func() {

		for {

			if session.IsClosed() {
				return
			}

			select {
			case msg := <-w.ToBrowser:
				s.Log(golive.LogDebug, "message out", golive.LogEx{"msg": msg, "session": sessionKey})
				if err := c.WriteJSON(msg); err != nil {
					s.Log(golive.LogError, "handle ws request: write json", golive.LogEx{"error": err})
				}
			case <-session.ExitSignal:
				if err := c.Close(); err != nil {
					s.Log(golive.LogError, "close websocket connection", golive.LogEx{"error": err})
				}
				// s.DeleteSession(sessionKey)
				s.Log(golive.LogInfo, "websocket close", golive.LogEx{"session": sessionKey})
				return
			}

		}
	}()

	for {
		if session.IsClosed() {
			return
		}

		inMsg := wire.Event{}

		// Loop blocks here
		if err := c.ReadJSON(&inMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				s.Log(golive.LogWarn, "handle ws request: unexpected connection close", nil)
				session.Close()
				return
			}

			if websocket.IsCloseError(err) {
				s.Log(golive.LogWarn, "handle ws request: unexpected close", nil)
				session.Close()
				return
			}

			s.Log(golive.LogError, "handle ws request: read json", golive.LogEx{"error": err})
			continue
		}
		s.Log(golive.LogDebug, "message in", golive.LogEx{"msg": inMsg, "session": sessionKey})

		w.HandleFromBrowser(&inMsg)
	}
}
