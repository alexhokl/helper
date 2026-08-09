package telemetry

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
)

func TestErrorLoggingUnaryInterceptor(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	original := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(original)

	tests := []struct {
		name          string
		handlerErr    error
		expectSuccess bool
	}{
		{"success", nil, true},
		{"failure", errors.New("boom"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, span := StartSpan(context.Background(), "test-tracer", test.name)

			handler := func(ctx context.Context, req any) (any, error) {
				return "ok", test.handlerErr
			}
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			resp, err := ErrorLoggingUnaryInterceptor(ctx, nil, info, handler)
			span.End()

			if !errors.Is(err, test.handlerErr) {
				t.Errorf("expected error %v but got %v", test.handlerErr, err)
			}
			if test.expectSuccess && resp != "ok" {
				t.Errorf("expected response %q but got %v", "ok", resp)
			}
		})
	}
}
