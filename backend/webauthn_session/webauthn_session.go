package webauthn_session

import (
	"encoding/gob"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type WebauthnSession struct {
	sess *sessions.Session
}

func NewWebauthnSession() *WebauthnSession {
	return &WebauthnSession{}
}

func (s *WebauthnSession) AddSessionMiddleware(e *echo.Echo) {
	// session準備
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	// sessionに登録する構造体を登録する。
	gob.Register(webauthn.SessionData{})
}

func (s *WebauthnSession) CreateSession(c echo.Context) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	}
	s.sess = sess
}

func (s *WebauthnSession) GetValue(key string) interface{} {
	return s.sess.Values[key]
}

func (s *WebauthnSession) SetValue(key string, value interface{}) {
	s.sess.Values[key] = value
}

func (s *WebauthnSession) Save(c echo.Context) error {
	return s.sess.Save(c.Request(), c.Response())
}
