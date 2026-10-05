package observability

import (
	"context"
	"log/slog"

	"example.com/app/internal/pkg/logging"
	"go.opentelemetry.io/otel/trace"
)

// correlationHandler appends request_id, trace_id and span_id to every record
// whose context carries them. It is the only way those keys reach a log line;
// call sites never pass them. The trace gate is IsValid, not IsSampled, so
// stderr and the OTLP bridge always agree on correlation.
type correlationHandler struct {
	next slog.Handler
}

func newCorrelationHandler(next slog.Handler) *correlationHandler {
	return &correlationHandler{next: next}
}

func (h *correlationHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *correlationHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := logging.RequestIDFromContext(ctx); ok {
		r.AddAttrs(slog.String(logging.KeyRequestID, id))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String(logging.KeyTraceID, sc.TraceID().String()),
			slog.String(logging.KeySpanID, sc.SpanID().String()),
		)
	}
	return h.next.Handle(ctx, r)
}

func (h *correlationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newCorrelationHandler(h.next.WithAttrs(attrs))
}

// WithGroup is passed through; this layout never opens slog groups, so
// correlation keys always land at the top level.
func (h *correlationHandler) WithGroup(name string) slog.Handler {
	return newCorrelationHandler(h.next.WithGroup(name))
}
