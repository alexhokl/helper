package telemetry

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

// mockHandler is a test handler that records calls
type mockHandler struct {
	records    []slog.Record
	attrs      []slog.Attr
	groupNames []string
	enabled    bool
}

func newMockHandler() *mockHandler {
	return &mockHandler{enabled: true}
}

func (h *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabled
}

func (h *mockHandler) Handle(ctx context.Context, record slog.Record) error {
	h.records = append(h.records, record)
	return nil
}

func (h *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &mockHandler{
		records:    h.records,
		attrs:      append(h.attrs, attrs...),
		groupNames: h.groupNames,
		enabled:    h.enabled,
	}
	return newHandler
}

func (h *mockHandler) WithGroup(name string) slog.Handler {
	newHandler := &mockHandler{
		records:    h.records,
		attrs:      h.attrs,
		groupNames: append(h.groupNames, name),
		enabled:    h.enabled,
	}
	return newHandler
}

func TestGCPLogHandlerConstants(t *testing.T) {
	// Verify constants are set correctly
	if GCPTraceIDKey != "logging.googleapis.com/trace" {
		t.Errorf("GCPTraceIDKey = %q, want %q", GCPTraceIDKey, "logging.googleapis.com/trace")
	}
	if GCPSpanIDKey != "logging.googleapis.com/spanId" {
		t.Errorf("GCPSpanIDKey = %q, want %q", GCPSpanIDKey, "logging.googleapis.com/spanId")
	}
	if GCPTraceSampledKey != "logging.googleapis.com/trace_sampled" {
		t.Errorf("GCPTraceSampledKey = %q, want %q", GCPTraceSampledKey, "logging.googleapis.com/trace_sampled")
	}
}

func TestNewGCPLogHandler(t *testing.T) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

	if handler == nil {
		t.Fatal("NewGCPLogHandler() returned nil")
	}

	gcpHandler, ok := handler.(GCPLogHandler)
	if !ok {
		t.Fatal("NewGCPLogHandler() did not return GCPLogHandler type")
	}

	if gcpHandler.Handler == nil {
		t.Error("NewGCPLogHandler() inner handler is nil")
	}
}

func TestGCPLogHandlerHandle(t *testing.T) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

	ctx := context.Background()
	record := slog.Record{}
	record.Message = "test message"

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("GCPLogHandler.Handle() error: %v", err)
	}

	if len(mockH.records) != 1 {
		t.Errorf("GCPLogHandler.Handle() recorded %d records, want 1", len(mockH.records))
	}
}

func TestGCPLogHandlerHandleWithNilHandler(t *testing.T) {
	handler := GCPLogHandler{Handler: nil}

	ctx := context.Background()
	record := slog.Record{}

	err := handler.Handle(ctx, record)
	if err == nil {
		t.Error("GCPLogHandler.Handle() with nil handler should return error")
	}
}

func TestGCPLogHandlerHandleWithSpanContext(t *testing.T) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

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
		t.Fatalf("GCPLogHandler.Handle() error: %v", err)
	}

	if len(mockH.records) != 1 {
		t.Fatalf("GCPLogHandler.Handle() recorded %d records, want 1", len(mockH.records))
	}

	// Check that trace attributes were added
	recordedRecord := mockH.records[0]
	var hasTraceID, hasSpanID, hasTraceSampled bool
	recordedRecord.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case GCPTraceIDKey:
			hasTraceID = true
		case GCPSpanIDKey:
			hasSpanID = true
		case GCPTraceSampledKey:
			hasTraceSampled = true
		}
		return true
	})

	if !hasTraceID {
		t.Error("GCPLogHandler.Handle() did not add trace ID attribute")
	}
	if !hasSpanID {
		t.Error("GCPLogHandler.Handle() did not add span ID attribute")
	}
	if !hasTraceSampled {
		t.Error("GCPLogHandler.Handle() did not add trace sampled attribute")
	}
}

func TestGCPLogHandlerWithAttrs(t *testing.T) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := handler.WithAttrs(attrs)
	if newHandler == nil {
		t.Fatal("GCPLogHandler.WithAttrs() returned nil")
	}

	_, ok := newHandler.(GCPLogHandler)
	if !ok {
		t.Error("GCPLogHandler.WithAttrs() did not return GCPLogHandler type")
	}
}

func TestGCPLogHandlerWithAttrsNilHandler(t *testing.T) {
	handler := GCPLogHandler{Handler: nil}

	newHandler := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})
	if newHandler == nil {
		t.Fatal("GCPLogHandler.WithAttrs() with nil handler returned nil")
	}
}

func TestGCPLogHandlerWithGroup(t *testing.T) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Fatal("GCPLogHandler.WithGroup() returned nil")
	}

	_, ok := newHandler.(GCPLogHandler)
	if !ok {
		t.Error("GCPLogHandler.WithGroup() did not return GCPLogHandler type")
	}
}

func TestGCPLogHandlerWithGroupNilHandler(t *testing.T) {
	handler := GCPLogHandler{Handler: nil}

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Fatal("GCPLogHandler.WithGroup() with nil handler returned nil")
	}
}

func TestGCPLogAttributeReplacer(t *testing.T) {
	tests := []struct {
		name    string
		attr    slog.Attr
		wantKey string
		wantVal string
	}{
		{
			name:    "level key renamed to severity",
			attr:    slog.Any(slog.LevelKey, slog.LevelInfo),
			wantKey: "severity",
			wantVal: "",
		},
		{
			name:    "warn level renamed to WARNING",
			attr:    slog.Any(slog.LevelKey, slog.LevelWarn),
			wantKey: "severity",
			wantVal: "WARNING",
		},
		{
			name:    "time key renamed to timestamp",
			attr:    slog.String(slog.TimeKey, "2024-01-01"),
			wantKey: "timestamp",
			wantVal: "",
		},
		{
			name:    "message key unchanged",
			attr:    slog.String(slog.MessageKey, "test message"),
			wantKey: "message",
			wantVal: "",
		},
		{
			name:    "other keys unchanged",
			attr:    slog.String("custom", "value"),
			wantKey: "custom",
			wantVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GCPLogAttributeReplacer(nil, tt.attr)
			if result.Key != tt.wantKey {
				t.Errorf("GCPLogAttributeReplacer() key = %q, want %q", result.Key, tt.wantKey)
			}
			if tt.wantVal != "" && result.Value.String() != tt.wantVal {
				t.Errorf("GCPLogAttributeReplacer() value = %q, want %q", result.Value.String(), tt.wantVal)
			}
		})
	}
}

func TestGCPLogHandlerIntegration(t *testing.T) {
	// Integration test with real slog.JSONHandler
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, nil)
	gcpHandler := NewGCPLogHandler(jsonHandler)

	logger := slog.New(gcpHandler)

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
		t.Error("GCPLogHandler integration test produced no output")
	}

	// Verify the output contains trace information
	if !bytes.Contains(buf.Bytes(), []byte(GCPTraceIDKey)) {
		t.Error("GCPLogHandler integration test output missing trace ID")
	}
}

// Benchmark tests
func BenchmarkGCPLogHandlerHandle(b *testing.B) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)
	ctx := context.Background()
	record := slog.Record{}
	record.Message = "benchmark message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, record)
	}
}

func BenchmarkGCPLogHandlerHandleWithSpan(b *testing.B) {
	mockH := newMockHandler()
	handler := NewGCPLogHandler(mockH)

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

func BenchmarkGCPLogAttributeReplacer(b *testing.B) {
	attr := slog.Any(slog.LevelKey, slog.LevelWarn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GCPLogAttributeReplacer(nil, attr)
	}
}
