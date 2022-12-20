package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var webAuthn *webauthn.WebAuthn
var userDB *userdb

// var sessionStore *session.Store

func main() {

	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Foobar Corp.",     // Display Name for your site
		RPID:          "localhost",        // Generally the domain name for your site
		RPOrigin:      "http://localhost", // The origin URL for WebAuthn requests
		// RPIcon: "https://duo.com/logo.png", // Optional icon URL for your site
	})

	if err != nil {
		log.Fatal("failed to create WebAuthn from config:", err)
	}

	userDB = DB()

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", index)

	e.GET("/register/begin/:username", BeginRegistration)
	// e.POST("/register/finish/{username}", FinishRegistration)
	// e.GET("/login/begin/{username}", BeginLogin)
	// e.POST("/login/finish/{username}", FinishLogin)

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
	// options, sessionData, err := webAuthn.BeginRegistration(
	options, _, err := webAuthn.BeginRegistration(
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
	// err = sessionStore.SaveWebauthnSession("registration", sessionData, r, w)
	// if err != nil {
	// 	log.Println(err)
	// 	jsonResponse(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	return jsonResponse(c, options, http.StatusOK)
}

func jsonResponse(c echo.Context, d interface{}, httpStatus int) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(httpStatus)
	return json.NewEncoder(c.Response()).Encode(d)
}
