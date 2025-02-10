package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	TraceIDKey = "trace_id"
	SpanIDKey  = "span_id"
)

// LogHandler is a log handler that sends log to OpenTelemetry.
type LogHandler struct {
	slog.Handler
}

// NewLogHandler returns a new log handler.
func NewLogHandler(handler slog.Handler) slog.Handler {
	return LogHandler{
		Handler: handler,
	}
}

// Handle logs the record to the log handler.
func (h LogHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.Handler == nil {
		return fmt.Errorf("log handler is not configured")
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		if spanCtx.HasTraceID() {
			record.AddAttrs(slog.String(TraceIDKey, spanCtx.TraceID().String()))
		}
		if spanCtx.HasSpanID() {
			record.AddAttrs(slog.String(SpanIDKey, spanCtx.SpanID().String()))
		}
	}

	return h.Handler.Handle(ctx, record)
}

// WithAttrs returns a new log handler with the given attributes.
func (h LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.Handler == nil {
		return h
	}
	return LogHandler{h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new log handler with the given group name.
func (h LogHandler) WithGroup(name string) slog.Handler {
	if h.Handler == nil {
		return h
	}
	return LogHandler{h.Handler.WithGroup(name)}
}
