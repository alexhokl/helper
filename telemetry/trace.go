package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a new child span with the given name using the tracer
// identified by tracerName.
func StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName)
}

// RecordSpanError records an error on the span and sets its status to Error.
func RecordSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// EndSpanOk marks the span status as Ok and ends it.
func EndSpanOk(span trace.Span) {
	span.SetStatus(codes.Ok, "")
	span.End()
}
