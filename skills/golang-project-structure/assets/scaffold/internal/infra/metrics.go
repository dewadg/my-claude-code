package infra

import (
	"context"
	"log/slog"

	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"example.com/app/internal/pkg/observability"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
)

// NewMeterProvider builds the metrics pipeline (no-op when off) and
// registers its flush.
func NewMeterProvider(lc fx.Lifecycle, cfg config.Config) (metric.MeterProvider, error) {
	mp, shutdown, err := observability.SetupMetrics(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, telemetryShutdownTimeout)
			defer cancel()

			if err := shutdown(ctx); err != nil {
				slog.With(logging.KeySource, "infra.NewMeterProvider").
					WarnContext(ctx, "metrics pipeline flush failed on shutdown", logging.KeyError, err)
			}
			return nil
		},
	})

	return mp, nil
}

// StartRuntimeMetrics starts Go runtime instrumentation. The logger is the
// first parameter on purpose: fx resolves left to right, so the configured
// logger is built (and its flush hook appended) before the meter provider.
func StartRuntimeMetrics(logger *slog.Logger, mp metric.MeterProvider) error {
	if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
		return err
	}
	logger.With(logging.KeySource, "infra.StartRuntimeMetrics").Debug("runtime metrics started")
	return nil
}
