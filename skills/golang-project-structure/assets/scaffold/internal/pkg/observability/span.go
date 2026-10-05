package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "example.com/app"

// Start opens a child span named "<package>.<Struct>.<Method>", the same
// string the method uses for its logging.KeySource.
//
//	func (s *Service) Get(ctx context.Context, id string) (_ Thing, err error) {
//		ctx, span := tracing.Start(ctx, "thing.Service.Get")
//		defer tracing.End(span, &err)
func Start(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name, trace.WithAttributes(attrs...))
}

// End records *err (when non-nil) on span, marks it errored, and ends it.
// Pass a pointer to the named error result so a defer sees the final value.
func End(span trace.Span, err *error) {
	if err != nil && *err != nil {
		span.RecordError(*err)
		span.SetStatus(codes.Error, (*err).Error())
	}
	span.End()
}

// Detach starts a new root span linked to the span in ctx, for work that
// outlives the request that triggered it. The returned context drops ctx's
// cancellation and values.
func Detach(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	if !trace.SpanContextFromContext(ctx).IsValid() {
		return tracer.Start(context.Background(), name)
	}
	return tracer.Start(context.Background(), name, trace.WithLinks(trace.LinkFromContext(ctx)))
}
