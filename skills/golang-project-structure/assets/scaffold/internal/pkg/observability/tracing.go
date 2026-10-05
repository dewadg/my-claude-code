package observability

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// SetupTracing builds the trace pipeline, or a no-op provider when off, and
// installs it plus the propagator as OTel globals. The propagator is set on
// both branches so header extraction never depends on export state. No
// sampler is set in code; OTEL_TRACES_SAMPLER stays an operator decision.
//
// It also routes SDK-internal errors (failed exports) through slog at WARN;
// left alone they surface as source-less INFO lines.
func SetupTracing(ctx context.Context, cfg config.Config) (trace.TracerProvider, func(context.Context) error, error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.With(logging.KeySource, "observability.otel").Warn("otel sdk error", logging.KeyError, err)
	}))
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.OTelTracesEnabled {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp, func(context.Context) error { return nil }, nil
	}

	setServiceIdentityEnv(cfg.Env)

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("observability: create OTLP trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)

	return tp, tp.Shutdown, nil
}
