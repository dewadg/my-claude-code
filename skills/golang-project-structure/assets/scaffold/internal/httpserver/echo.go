package httpserver

import (
	"github.com/labstack/echo/v4"
)

// NewEcho builds the router with banner and port logging suppressed.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	return e
}
