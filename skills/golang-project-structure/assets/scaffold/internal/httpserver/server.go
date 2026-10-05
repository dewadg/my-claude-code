package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// NewHTTPServer is pure: it registers no hooks.
func NewHTTPServer(cfg config.Config, e *echo.Echo) *http.Server {
	return &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

// RegisterLifecycle binds the socket synchronously in OnStart, so a port in
// use fails startup, then serves on a goroutine. A serve error shuts the app
// down with exit code 1.
func RegisterLifecycle(lc fx.Lifecycle, sh fx.Shutdowner, srv *http.Server, cfg config.Config) {
	l := slog.With(logging.KeySource, "httpserver.RegisterLifecycle")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			l.InfoContext(ctx, "server started", "port", cfg.Port)
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					l.Error("server failed", logging.KeyError, err)
					_ = sh.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			l.InfoContext(ctx, "server stopping")
			return srv.Shutdown(ctx)
		},
	})
}
