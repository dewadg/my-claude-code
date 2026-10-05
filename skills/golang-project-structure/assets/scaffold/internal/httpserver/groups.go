package httpserver

import (
	"github.com/labstack/echo/v4"
)

// NewPublicGroup is the unauthenticated /api group.
func NewPublicGroup(e *echo.Echo) *echo.Group {
	return e.Group("/api")
}

// NewPrivateGroup is the authenticated /api group. Attach auth/CORS
// middleware here, once, so no domain can mount a private route without it.
func NewPrivateGroup(e *echo.Echo) *echo.Group {
	return e.Group("/api")
}
