// Package infra provides process-level dependencies every domain builds on:
// the configured *slog.Logger, the OTel meter and tracer providers, and the
// database handle. It is not a domain: no routes, no service layering.
//
// Whoever opens a resource registers its OnStop next to the open. fx runs
// OnStop hooks LIFO in construction order, so the HTTP server drains first,
// then traces flush, then the database closes, then metrics and logs flush.
// Telemetry flush failures are logged, never returned: a returned OnStop
// error turns a graceful SIGTERM into a non-zero exit.
package infra
