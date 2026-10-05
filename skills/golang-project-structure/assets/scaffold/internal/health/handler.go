package health

import (
	"net/http"

	tracing "example.com/app/internal/pkg/observability"
	"github.com/labstack/echo/v4"
)

type handlers struct {
	svc healthService
}

// handleGet answers 200 when every check passes, 503 otherwise.
func (h *handlers) handleGet(c echo.Context) (err error) {
	ctx, span := tracing.Start(c.Request().Context(), "health.Handler.Get")
	defer tracing.End(span, &err)

	report := h.svc.Check(ctx)

	status := http.StatusOK
	if report.Status != StatusOK {
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, report)
}
