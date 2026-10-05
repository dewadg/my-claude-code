package logging

import "context"

type requestIDKey struct{}

// WithRequestID returns a copy of ctx carrying the request correlation id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the id stored by WithRequestID; ok is false
// when none or an empty one was stored.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id, id != ""
}
