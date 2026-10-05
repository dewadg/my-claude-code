package logging

// Canonical slog attribute keys. Call sites reference these instead of
// repeating string literals.
const (
	// KeySource names the emitting component as "<package>.<Struct>.<Method>",
	// identical to the span name opened by the same method.
	KeySource = "source"
	// KeyError carries an error value. Never "err".
	KeyError = "error"
	// KeyRequestID, KeyTraceID and KeySpanID are injected by the correlation
	// handler; call sites never pass them.
	KeyRequestID = "request_id"
	KeyTraceID   = "trace_id"
	KeySpanID    = "span_id"
	// KeyDurationMS carries an elapsed duration in milliseconds.
	KeyDurationMS = "duration_ms"
	// KeyCount carries an item count. Never "len".
	KeyCount = "count"
)
