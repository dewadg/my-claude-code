// Package observability builds the three OpenTelemetry pipelines, one entry
// point each: SetupLogs, SetupMetrics, SetupTracing. Each is gated by its own
// config flag and returns a working local substitute when off (stderr-only
// logger, no-op meter provider, no-op tracer provider), so no instrumentation
// call site ever branches on telemetry being enabled.
//
// Exporters are OTLP/HTTP and read OTEL_EXPORTER_OTLP_ENDPOINT,
// OTEL_EXPORTER_OTLP_HEADERS, OTEL_TRACES_SAMPLER and
// OTEL_METRIC_EXPORT_INTERVAL from the process env through the SDK.
//
// SetupTracing installs the global tracer provider and a TraceContext+Baggage
// propagator; Start/End/Detach in span.go read that global so services and
// repositories open spans without taking a TracerProvider parameter.
package observability
