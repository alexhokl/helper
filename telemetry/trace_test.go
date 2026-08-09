package telemetry

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// newRecordingTracerProvider returns a TracerProvider recording spans into
// the returned exporter so span status and events can be asserted.
func newRecordingTracerProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	return sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)), recorder
}

func TestSpanHelpers(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus codes.Code
	}{
		{"ok", nil, codes.Ok},
		{"error", errors.New("boom"), codes.Error},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tp, recorder := newRecordingTracerProvider(t)
			original := otel.GetTracerProvider()
			otel.SetTracerProvider(tp)
			defer otel.SetTracerProvider(original)

			ctx, span := StartSpan(context.Background(), "test-tracer", test.name)
			if !span.SpanContext().IsValid() {
				t.Fatalf("expected a valid span context")
			}
			if trace.SpanFromContext(ctx) != span {
				t.Errorf("expected span to be stored in the returned context")
			}

			if test.err != nil {
				RecordSpanError(span, test.err)
				span.End()
			} else {
				EndSpanOk(span)
			}

			ended := recorder.Ended()
			if len(ended) != 1 {
				t.Fatalf("expected 1 ended span but got %d", len(ended))
			}
			if ended[0].Status().Code != test.expectedStatus {
				t.Errorf("expected status %v but got %v", test.expectedStatus, ended[0].Status().Code)
			}
		})
	}
}
