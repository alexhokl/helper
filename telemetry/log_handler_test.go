package telemetry

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestLogHandlerConstants(t *testing.T) {
	// Verify constants are set correctly
	if TraceIDKey != "trace_id" {
		t.Errorf("TraceIDKey = %q, want %q", TraceIDKey, "trace_id")
	}
	if SpanIDKey != "span_id" {
		t.Errorf("SpanIDKey = %q, want %q", SpanIDKey, "span_id")
	}
}

func TestNewLogHandler(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	if handler == nil {
		t.Fatal("NewLogHandler() returned nil")
	}

	logHandler, ok := handler.(LogHandler)
	if !ok {
		t.Fatal("NewLogHandler() did not return LogHandler type")
	}

	if logHandler.Handler == nil {
		t.Error("NewLogHandler() inner handler is nil")
	}
}

func TestLogHandlerHandle(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	ctx := context.Background()
	record := slog.Record{}
	record.Message = "test message"

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("LogHandler.Handle() error: %v", err)
	}

	if len(mockH.records) != 1 {
		t.Errorf("LogHandler.Handle() recorded %d records, want 1", len(mockH.records))
	}
}

func TestLogHandlerHandleWithNilHandler(t *testing.T) {
	handler := LogHandler{Handler: nil}

	ctx := context.Background()
	record := slog.Record{}

	err := handler.Handle(ctx, record)
	if err == nil {
		t.Error("LogHandler.Handle() with nil handler should return error")
	}
}

func TestLogHandlerHandleWithSpanContext(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	// Create a valid span context
	traceID, _ := trace.TraceIDFromHex("00000000000000000000000000000001")
	spanID, _ := trace.SpanIDFromHex("0000000000000001")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
	record := slog.Record{}
	record.Message = "test with span"

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("LogHandler.Handle() error: %v", err)
	}

	if len(mockH.records) != 1 {
		t.Fatalf("LogHandler.Handle() recorded %d records, want 1", len(mockH.records))
	}

	// Check that trace attributes were added
	recordedRecord := mockH.records[0]
	var hasTraceID, hasSpanID bool
	recordedRecord.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case TraceIDKey:
			hasTraceID = true
		case SpanIDKey:
			hasSpanID = true
		}
		return true
	})

	if !hasTraceID {
		t.Error("LogHandler.Handle() did not add trace ID attribute")
	}
	if !hasSpanID {
		t.Error("LogHandler.Handle() did not add span ID attribute")
	}
}

func TestLogHandlerHandleWithInvalidSpanContext(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	// Create an invalid (empty) span context
	ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContext{})
	record := slog.Record{}
	record.Message = "test with invalid span"

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("LogHandler.Handle() error: %v", err)
	}

	// Check that no trace attributes were added for invalid span
	recordedRecord := mockH.records[0]
	var hasTraceID, hasSpanID bool
	recordedRecord.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case TraceIDKey:
			hasTraceID = true
		case SpanIDKey:
			hasSpanID = true
		}
		return true
	})

	if hasTraceID {
		t.Error("LogHandler.Handle() should not add trace ID for invalid span context")
	}
	if hasSpanID {
		t.Error("LogHandler.Handle() should not add span ID for invalid span context")
	}
}

func TestLogHandlerWithAttrs(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := handler.WithAttrs(attrs)
	if newHandler == nil {
		t.Fatal("LogHandler.WithAttrs() returned nil")
	}

	_, ok := newHandler.(LogHandler)
	if !ok {
		t.Error("LogHandler.WithAttrs() did not return LogHandler type")
	}
}

func TestLogHandlerWithAttrsNilHandler(t *testing.T) {
	handler := LogHandler{Handler: nil}

	newHandler := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})
	if newHandler == nil {
		t.Fatal("LogHandler.WithAttrs() with nil handler returned nil")
	}
}

func TestLogHandlerWithGroup(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Fatal("LogHandler.WithGroup() returned nil")
	}

	_, ok := newHandler.(LogHandler)
	if !ok {
		t.Error("LogHandler.WithGroup() did not return LogHandler type")
	}
}

func TestLogHandlerWithGroupNilHandler(t *testing.T) {
	handler := LogHandler{Handler: nil}

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Fatal("LogHandler.WithGroup() with nil handler returned nil")
	}
}

func TestLogHandlerIntegration(t *testing.T) {
	// Integration test with real slog.JSONHandler
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, nil)
	logHandler := NewLogHandler(jsonHandler)

	logger := slog.New(logHandler)

	// Create a span context
	traceID, _ := trace.TraceIDFromHex("00000000000000000000000000000001")
	spanID, _ := trace.SpanIDFromHex("0000000000000001")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	logger.InfoContext(ctx, "test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Error("LogHandler integration test produced no output")
	}

	// Verify the output contains trace information
	if !bytes.Contains(buf.Bytes(), []byte(TraceIDKey)) {
		t.Error("LogHandler integration test output missing trace ID")
	}
	if !bytes.Contains(buf.Bytes(), []byte(SpanIDKey)) {
		t.Error("LogHandler integration test output missing span ID")
	}
}

func TestLogHandlerChaining(t *testing.T) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	// Chain multiple operations
	handler = handler.WithAttrs([]slog.Attr{slog.String("attr1", "val1")}).(LogHandler)
	handler = handler.WithGroup("group1").(LogHandler)
	handler = handler.WithAttrs([]slog.Attr{slog.String("attr2", "val2")}).(LogHandler)

	ctx := context.Background()
	record := slog.Record{}
	record.Message = "chained test"

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("LogHandler chained Handle() error: %v", err)
	}
}

// Benchmark tests
func BenchmarkLogHandlerHandle(b *testing.B) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)
	ctx := context.Background()
	record := slog.Record{}
	record.Message = "benchmark message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, record)
	}
}

func BenchmarkLogHandlerHandleWithSpan(b *testing.B) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)

	traceID, _ := trace.TraceIDFromHex("00000000000000000000000000000001")
	spanID, _ := trace.SpanIDFromHex("0000000000000001")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
	record := slog.Record{}
	record.Message = "benchmark message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, record)
	}
}

func BenchmarkLogHandlerWithAttrs(b *testing.B) {
	mockH := newMockHandler()
	handler := NewLogHandler(mockH)
	attrs := []slog.Attr{slog.String("key", "value")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.WithAttrs(attrs)
	}
}
