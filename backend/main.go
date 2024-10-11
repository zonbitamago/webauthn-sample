package main

// 以下を参考に実装
// https://github.com/hbolimovsky/webauthn-example/blob/master/server.go

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"webauthn-sample-backend/webauthn_session"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var webAuthn *webauthn.WebAuthn
var userDB *userdb
var webauthnSession *webauthn_session.WebauthnSession

func main() {

	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Foobar Corp.",                               // Display Name for your site
		RPID:          "star-united-panther.ngrok-free.app",         // Generally the domain name for your site
		RPOrigin:      "https://star-united-panther.ngrok-free.app", // The origin URL for WebAuthn requests
		// RPIcon: "https://duo.com/logo.png", // Optional icon URL for your site
	})

	if err != nil {
		log.Fatal("failed to create WebAuthn from config:", err)
	}

	userDB = DB()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS設定
	e.Use(middleware.CORSWithConfig(
		middleware.CORSConfig{
			// Method
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPut,
				http.MethodPost,
				http.MethodDelete,
				http.MethodOptions,
			},
			// Origin
			AllowOrigins: []string{
				"https://star-united-panther.ngrok-free.app",
			},
			// CORS
			AllowCredentials: true,
		}))

	// session準備
	webauthnSession = webauthn_session.NewWebauthnSession()
	webauthnSession.AddSessionMiddleware(e)

	e.GET("/api/", index)

	e.GET("/api/register/begin/:username", BeginRegistration)
	e.POST("/api/register/finish/:username", FinishRegistration)
	e.GET("/api/login/begin/:username", BeginLogin)
	e.POST("/api/login/finish/:username", FinishLogin)

	port := 1323
	e.Logger.Info(fmt.Sprintf("ServerStartUp! port:%v", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}

func index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func BeginRegistration(c echo.Context) error {
	webauthnSession.CreateSession(c)

	// get username/friendly name
	username := c.Param("username")
	if username == "" {
		er := &ErrorResponse{
			Message: "must supply a valid username i.e. foo@bar.com",
		}
		return jsonResponse(c, er, http.StatusBadRequest)
	}

	// get user
	user, err := userDB.GetUser(username)
	// user doesn't exist, create new user
	if err != nil {
		displayName := strings.Split(username, "@")[0]
		user = NewUser(username, displayName)
		userDB.PutUser(user)
	}

	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = user.CredentialExcludeList()
	}

	// generate PublicKeyCredentialCreationOptions, session data
	options, sessionData, err := webAuthn.BeginRegistration(
		user,
		registerOptions,
	)

	if err != nil {
		c.Logger().Error(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusInternalServerError)
	}

	// store session data as marshaled JSON
	webauthnSession.SetValue("registration", sessionData)
	err = webauthnSession.Save(c)
	if err != nil {
		c.Logger().Error(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusInternalServerError)
	}

	return jsonResponse(c, options, http.StatusOK)
}

func FinishRegistration(c echo.Context) error {
	webauthnSession.CreateSession(c)

	// get username
	username := c.Param("username")

	// get user
	user, err := userDB.GetUser(username)

	// user doesn't exist
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusBadRequest)
	}

	// load the session data
	sessionData := webauthnSession.GetValue("registration").(webauthn.SessionData)

	credential, err := webAuthn.FinishRegistration(user, sessionData, c.Request())
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusBadRequest)
	}

	user.AddCredential(*credential)

	return jsonResponse(c, "Registration Success", http.StatusOK)
}

func BeginLogin(c echo.Context) error {
	webauthnSession.CreateSession(c)

	// get username
	username := c.Param("username")

	// get user
	user, err := userDB.GetUser(username)

	// user doesn't exist
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusBadRequest)
	}

	// generate PublicKeyCredentialRequestOptions, session data
	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusInternalServerError)
	}

	// store session data as marshaled JSON

	webauthnSession.SetValue("authentication", sessionData)
	webauthnSession.Save(c)

	return jsonResponse(c, options, http.StatusOK)
}

func FinishLogin(c echo.Context) error {
	webauthnSession.CreateSession(c)

	// get username
	username := c.Param("username")

	// get user
	user, err := userDB.GetUser(username)

	// user doesn't exist
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusBadRequest)
	}

	// load the session data
	sessionData := webauthnSession.GetValue("authentication").(webauthn.SessionData)

	// in an actual implementation, we should perform additional checks on
	// the returned 'credential', i.e. check 'credential.Authenticator.CloneWarning'
	// and then increment the credentials counter
	_, err = webAuthn.FinishLogin(user, sessionData, c.Request())
	if err != nil {
		log.Println(err)
		er := &ErrorResponse{
			Message: err.Error(),
		}

		return jsonResponse(c, er, http.StatusBadRequest)
	}

	// handle successful login
	return jsonResponse(c, "Login Success", http.StatusOK)
}

func jsonResponse(c echo.Context, d interface{}, httpStatus int) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(httpStatus)
	return json.NewEncoder(c.Response()).Encode(d)
}
