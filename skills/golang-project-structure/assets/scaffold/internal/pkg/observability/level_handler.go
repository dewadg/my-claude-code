package observability

import (
	"context"
	"log/slog"
)

// levelHandler gates a handler at a fixed minimum level. The OTel bridge's
// own Enabled defers to the LoggerProvider, which has no minimum, so without
// this gate every level would be exported regardless of LOG_LEVEL.
type levelHandler struct {
	next slog.Handler
	min  slog.Level
}

func newLevelHandler(next slog.Handler, min slog.Level) *levelHandler {
	return &levelHandler{next: next, min: min}
}

func (h *levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.min && h.next.Enabled(ctx, level)
}

func (h *levelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.next.Handle(ctx, r)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newLevelHandler(h.next.WithAttrs(attrs), h.min)
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return newLevelHandler(h.next.WithGroup(name), h.min)
}
