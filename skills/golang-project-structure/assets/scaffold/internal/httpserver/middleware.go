package httpserver

import (
	"log/slog"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/logging"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
)

// RegisterMiddleware installs, outermost first: request id (bridged into
// context.Context for the correlation handler), the OTel server span, panic
// recovery, and the access log. Taking tp is what makes fx build the tracer
// provider.
func RegisterMiddleware(e *echo.Echo, log *slog.Logger, tp trace.TracerProvider) {
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, id string) {
			c.SetRequest(c.Request().WithContext(logging.WithRequestID(c.Request().Context(), id)))
		},
	}))
	e.Use(otelecho.Middleware(buildutil.Name, otelecho.WithTracerProvider(tp)))
	e.Use(middleware.Recover())

	access := log.With(logging.KeySource, "httpserver.AccessLog")
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:  true,
		LogMethod:    true,
		LogURIPath:   true,
		LogRoutePath: true,
		LogStatus:    true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			if v.Status >= 500 {
				level = slog.LevelError
			}
			access.Log(c.Request().Context(), level, "request completed",
				"method", v.Method,
				"route", v.RoutePath,
				"path", v.URIPath,
				"status", v.Status,
				logging.KeyDurationMS, v.Latency.Milliseconds(),
			)
			return nil
		},
	}))
}
