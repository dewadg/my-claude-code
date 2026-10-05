package infra

import (
	"context"
	"log/slog"
	"os"
	"time"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"example.com/app/internal/pkg/observability"
	"go.uber.org/fx"
)

// telemetryShutdownTimeout caps each pipeline's final flush so a dead
// collector cannot eat fx's whole stop budget.
const telemetryShutdownTimeout = 5 * time.Second

// BaseLoggerOut tags the stderr-only logger "base" so it does not collide
// with the final *slog.Logger.
type BaseLoggerOut struct {
	fx.Out

	Base *slog.Logger `name:"base"`
}

// NewBaseLogger builds the stderr logger from LOG_LEVEL and LOG_FORMAT, with
// a "service" attribute on every record.
func NewBaseLogger(cfg config.Config) BaseLoggerOut {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}

	var h slog.Handler
	if cfg.LogFormat == "json" {
		h = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		h = slog.NewTextHandler(os.Stderr, opts)
	}

	return BaseLoggerOut{Base: slog.New(h).With("service", buildutil.Name)}
}

// NewLogger extends base with correlation and optional OTLP export, and
// installs it as slog's default so package-level slog calls everywhere use
// it. The flush warning goes to base: by then the default logger fans out
// into the provider being shut down.
func NewLogger(lc fx.Lifecycle, cfg config.Config, base *slog.Logger) (*slog.Logger, error) {
	logger, shutdown, err := observability.SetupLogs(context.Background(), cfg, base)
	if err != nil {
		return nil, err
	}

	slog.SetDefault(logger)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, telemetryShutdownTimeout)
			defer cancel()

			if err := shutdown(ctx); err != nil {
				base.With(logging.KeySource, "infra.NewLogger").
					WarnContext(ctx, "log pipeline flush failed on shutdown", logging.KeyError, err)
			}
			return nil
		},
	})

	return logger, nil
}
