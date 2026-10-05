package httpapi

import (
	"log/slog"

	"example.com/app/internal/pkg/logging"
	"github.com/labstack/echo/v4"
)

// Envelope is the success body shape: {"data": ...}.
type Envelope struct {
	Data any `json:"data"`
}

// Data writes v as {"data": v}.
func Data(c echo.Context, status int, v any) error {
	return c.JSON(status, Envelope{Data: v})
}

// FailWith maps err through the domain's mapper and logs only when the
// result is a 5xx — 4xx are expected outcomes and are not logged. This is the
// single place a handler logs a failure, so errors are logged once, at the
// boundary.
func FailWith(c echo.Context, l *slog.Logger, msg string, err error, mapper func(error) error) error {
	mapped := mapper(err)
	if he, ok := mapped.(*echo.HTTPError); !ok || he.Code >= 500 {
		l.ErrorContext(c.Request().Context(), msg, logging.KeyError, err)
	}
	return mapped
}
