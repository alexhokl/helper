package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	GCPTraceIDKey      = "logging.googleapis.com/trace"
	GCPSpanIDKey       = "logging.googleapis.com/spanId"
	GCPTraceSampledKey = "logging.googleapis.com/trace_sampled"
)

// GCPLogHandler is a log handler that sends log to OpenTelemetry.
type GCPLogHandler struct {
	slog.Handler
}

// NewLogHandler returns a new log handler.
func NewGCPLogHandler(handler slog.Handler) slog.Handler {
	return GCPLogHandler{
		Handler: handler,
	}
}

// Handle logs the record to the log handler.
func (h GCPLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.Handler == nil {
		return fmt.Errorf("log handler is not configured")
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		if spanCtx.HasTraceID() {
			record.AddAttrs(slog.String(GCPTraceIDKey, spanCtx.TraceID().String()))
		}
		if spanCtx.HasSpanID() {
			record.AddAttrs(slog.String(GCPSpanIDKey, spanCtx.SpanID().String()))
		}
		if spanCtx.IsSampled() {
			record.AddAttrs(slog.Bool(GCPTraceSampledKey, spanCtx.TraceFlags().IsSampled()))
		}
	}

	return h.Handler.Handle(ctx, record)
}

// WithAttrs returns a new log handler with the given attributes.
func (h GCPLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.Handler == nil {
		return h
	}
	return GCPLogHandler{h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new log handler with the given group name.
func (h GCPLogHandler) WithGroup(name string) slog.Handler {
	if h.Handler == nil {
		return h
	}
	return GCPLogHandler{h.Handler.WithGroup(name)}
}

func GCPLogAttributeReplacer(groups []string, a slog.Attr) slog.Attr {
	// Rename attribute keys to match Cloud Logging structured log format
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		// Map slog.Level string values to Cloud Logging LogSeverity
		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		if level := a.Value.Any().(slog.Level); level == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.TimeKey:
		a.Key = "timestamp"
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}
