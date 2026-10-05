package observability

import (
	"context"
	"fmt"

	"example.com/app/internal/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// SetupMetrics builds the metrics pipeline, or a no-op provider when off.
func SetupMetrics(ctx context.Context, cfg config.Config) (metric.MeterProvider, func(context.Context) error, error) {
	if !cfg.OTelMetricsEnabled {
		mp := noop.NewMeterProvider()
		otel.SetMeterProvider(mp)
		return mp, func(context.Context) error { return nil }, nil
	}

	setServiceIdentityEnv(cfg.Env)

	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("observability: create OTLP metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	otel.SetMeterProvider(mp)

	return mp, mp.Shutdown, nil
}
