package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// ErrorLoggingUnaryInterceptor logs any error returned by the handler. It
// also records the error on the active OpenTelemetry span and sets its
// status to Error.
func ErrorLoggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		RecordSpanError(trace.SpanFromContext(ctx), err)
		slog.ErrorContext(
			ctx,
			"gRPC error",
			slog.String("method", info.FullMethod),
			slog.String("error", err.Error()),
		)
	}
	return resp, err
}
