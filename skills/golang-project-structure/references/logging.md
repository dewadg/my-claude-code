# Logging conventions (log/slog)

`log/slog` is the only logger. zap/zerolog/logrus and stdlib `log` are not used; `fmt.Print*` only
in `cmd/` for user-facing CLI output (e.g. `migrate status`).

## Handler chain (built once, in infra)

```
NewBaseLogger   stderr, LOG_LEVEL, LOG_FORMAT=text|json, With("service", buildutil.Name)   [name:"base"]
   └─ NewLogger → observability.SetupLogs
        correlationHandler            adds request_id, trace_id, span_id from ctx
          └─ levelHandler(LOG_LEVEL)  only when OTLP logs on: gates the fan-out
               └─ slogmulti.Fanout(base handler, otelslog bridge)
      slog.SetDefault(logger)
```

`slog.SetDefault` is deliberate: package-level `slog.With(...)` at call time then picks up the
configured handler without threading `*slog.Logger` through every constructor. Code that runs at
**construction** time and captures a logger takes `*slog.Logger` as an fx parameter instead (forces
fx to build the configured logger first).

## Call-site shape

```go
func (s *Service) Cancel(ctx context.Context, id string) (err error) {
	ctx, span := tracing.Start(ctx, "order.Service.Cancel")
	defer tracing.End(span, &err)

	l := slog.With(logging.KeySource, "order.Service.Cancel")

	if err := s.store.MarkCancelled(ctx, id); err != nil {
		return fmt.Errorf("mark cancelled: %w", err)    // propagate; caller logs
	}
	if err := s.events.Publish(ctx, id); err != nil {
		l.WarnContext(ctx, "publish cancel event failed", logging.KeyError, err, "order_id", id)  // handled here → logged here
	}
	l.InfoContext(ctx, "order cancelled", "order_id", id)
	return nil
}
```

## Rules

| Rule | Why |
| --- | --- |
| `l := slog.With(logging.KeySource, "<package>.<Struct>.<Method>")` at the top of the method; never package-level `slog.Info(...)` | every line is attributable; same string as the span name, so logs ↔ traces join |
| Use `logging.KeySource`, never the literal `"source"` | one place to rename |
| Always the `*Context` variants (`InfoContext(ctx, ...)`) when a ctx exists | the correlation handler reads request_id / trace_id / span_id from ctx |
| Never pass `request_id`, `trace_id`, `span_id` yourself | the correlation handler injects them; passing them duplicates keys |
| Error attr is `logging.KeyError` (`"error"`), never `"err"` | consistent query field |
| Message: short lowercase phrase, past tense or `<thing> failed`, no trailing punctuation, no interpolated values (`"order cancelled"`, `"publish cancel event failed"`) | messages group/aggregate; variables go in attrs |
| Attrs are key/value pairs with snake_case keys; shared keys are constants in `internal/pkg/logging/keys.go` (`KeyCount` not `"len"`, `KeyDurationMS`) | stable schema across services |
| No `WithGroup` | correlation keys must stay top-level for queries |
| **Log once, where the error is handled.** Returning it → don't log. Handler boundary → `httpapi.FailWith(c, l, "<op> failed", err, mapError)` logs only 5xx | no duplicate stack of the same failure |
| Levels: `Debug` dev detail · `Info` state change worth an audit trail · `Warn` degraded but handled (fallback, best-effort side effect failed) · `Error` request/job failed and needs attention | alert on Error, dashboard Warn |
| Expected outcomes (4xx, validation, not found) are not logged as errors | noise hides real failures |
| Never log secrets, tokens, passwords, full request bodies, or raw DSNs | logs are exported off-box |

## Config-time logging

Before `infra` runs, `main.go` installs a bootstrap text logger on stderr. `config` warnings
(`invalid duration env, using default`) go through it, still with a source
(`slog.With(logging.KeySource, "config.parseDuration")`).

## Shutdown

Telemetry flush failures log at `Warn` and return nil. The log pipeline's own flush warning goes
to the `base` logger (straight to stderr) because the default logger fans out into the provider
being shut down.

## Audit coverage

`scripts/audit.sh` WARNs on: package-level `slog.X(` calls, literal `"source"`, `"err"` keys,
uppercase messages, stdlib `log`, `fmt.Print*` outside `cmd/`; FAILs on third-party loggers and on
infra never calling `slog.SetDefault`.
