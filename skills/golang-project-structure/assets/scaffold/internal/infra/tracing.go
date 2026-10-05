package infra

import (
	"context"
	"log/slog"

	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"example.com/app/internal/pkg/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// NewTracerProvider builds the trace pipeline (no-op when off) and registers
// its flush. It is provided, not invoked: the HTTP middleware consumes it.
func NewTracerProvider(lc fx.Lifecycle, cfg config.Config) (trace.TracerProvider, error) {
	tp, shutdown, err := observability.SetupTracing(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, telemetryShutdownTimeout)
			defer cancel()

			if err := shutdown(ctx); err != nil {
				slog.With(logging.KeySource, "infra.NewTracerProvider").
					WarnContext(ctx, "trace pipeline flush failed on shutdown", logging.KeyError, err)
			}
			return nil
		},
	})

	return tp, nil
}
