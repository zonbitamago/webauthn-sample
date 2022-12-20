package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", index)

	port := 1323
	e.Logger.Info(fmt.Sprintf("ServerStartUp! port:%v", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}

func index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}
