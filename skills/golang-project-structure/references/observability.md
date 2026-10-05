# Observability, lifecycle, Docker and Makefile notes

## OTel pipelines

| Signal | Entry point | Off | On |
| --- | --- | --- | --- |
| logs | `observability.SetupLogs(ctx, cfg, base)` | base handler + correlation | + `otlploghttp` batch exporter via `otelslog` bridge, level-gated |
| metrics | `observability.SetupMetrics(ctx, cfg)` | `noop.NewMeterProvider()` | `otlpmetrichttp` + periodic reader |
| traces | `observability.SetupTracing(ctx, cfg)` | `noop.NewTracerProvider()` | `otlptracehttp` + batcher |

- All exporters are OTLP/HTTP and read `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`,
  `OTEL_TRACES_SAMPLER(_ARG)`, `OTEL_METRIC_EXPORT_INTERVAL` from the env through the SDK — never
  configured in code.
- `setServiceIdentityEnv` defaults `OTEL_SERVICE_NAME` (= `buildutil.Name`) and
  `OTEL_RESOURCE_ATTRIBUTES` (`service.version`, `deployment.environment.name`) **before** each SDK
  provider: `resource.Default()` is memoized, the first provider freezes it.
- `SetupTracing` sets global tracer provider + TraceContext/Baggage propagator (both branches) and
  `otel.SetErrorHandler` → slog WARN. Span helpers read the global, so domains never take a
  `TracerProvider`.
- HTTP: `otelecho.Middleware(buildutil.Name, otelecho.WithTracerProvider(tp))` after the
  request-id middleware (which bridges the id into `context.Context`). Taking `tp` as a
  `RegisterMiddleware` param is what makes fx build the tracer provider.
- `Detach(ctx, name)` for work outliving the request: new root span linked to the request span.

## fx wiring in infra

```go
fx.Provide(NewBaseLogger)                                          // fx.Out, name:"base"
fx.Provide(fx.Annotate(NewLogger, fx.ParamTags(``, ``, `name:"base"`)))
fx.Provide(NewMeterProvider)
fx.Provide(NewTracerProvider)
fx.Provide(NewDB)                                                  // lazy: first consumer builds it
fx.Invoke(StartRuntimeMetrics)                                     // (logger, mp) — logger first on purpose
```

OnStop runs LIFO in construction order → HTTP server drains → traces flush → DB closes → metrics
flush → logs flush. Each telemetry flush is capped at 5s and returns nil on failure. Whole-app stop
budget: `fx.StopTimeout(cfg.ShutdownTimeout)` (`SHUTDOWN_TIMEOUT`, default 15s).

Verified behavior: with `OTEL_ENABLED=true` and a dead collector the service still serves, access
logs carry `trace_id`/`span_id`, and SIGINT exits 0 after logging
`metrics pipeline flush failed on shutdown` / `log pipeline flush failed on shutdown` at WARN.

## Dockerfile shape

1. `golang:<ver>-alpine` builder: `COPY go.mod go.sum` → `go mod download` → `COPY cmd internal
   build` (build/ carries the embedded migrations — forgetting it breaks `//go:embed`) →
   `ARG BUILD_VERSION` (after sources, keeps the module layer cached) →
   `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X '<module>/internal/pkg/buildutil.Version=${BUILD_VERSION}'"`.
2. Extra fetch stages (CLIs, data files) as their own `FROM alpine` stages with pinned checksums.
3. Runtime `gcr.io/distroless/static-debian12:nonroot` (`cc-debian12` if cgo/glibc needed);
   `ENTRYPOINT ["/<bin>"]`, `CMD ["serve"]`. The same image runs `migrate` and `seed` as one-off
   containers before the deploy step. Writable volumes: pre-create the dir in the builder and
   `COPY --chown=65532:65532`.
4. `.dockerignore` excludes `.git`, `.env`, `bin/`, local tooling dirs.

## Makefile targets

| Target | Does |
| --- | --- |
| `dev` | `migrate seed serve` |
| `serve` | `go run ./cmd/<bin> serve` |
| `migrate` / `migrate-down` / `migrate-status` | goose up / down one / status, via the binary |
| `migrate-new name=x` | next sequential `build/migrations/NNNNN_x.sql` |
| `seed` | idempotent seeders |
| `build` / `run` | static binary in `bin/` with version ldflags / build then serve |
| `test` / `vet` / `lint` / `tidy` | `go test -race ./...` / `go vet` / `golangci-lint run` / `go mod tidy` |
| `docker-build` | `docker build --load --build-arg BUILD_VERSION=$(VERSION)` |

All targets read `./.env` through `config.LoadDotenv`; process env wins over `.env`.
