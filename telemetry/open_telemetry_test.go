package telemetry

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func TestNewResource(t *testing.T) {
	serviceName := "test-service"
	serviceVersion := "1.0.0"

	res := NewResource(serviceName, serviceVersion)

	if res == nil {
		t.Fatal("NewResource() returned nil")
	}

	// Check that attributes are set
	attrs := res.Attributes()
	var hasServiceName, hasServiceVersion, hasHostName bool

	for _, attr := range attrs {
		switch attr.Key {
		case semconv.ServiceNameKey:
			hasServiceName = true
			if attr.Value.AsString() != serviceName {
				t.Errorf("NewResource() service name = %q, want %q", attr.Value.AsString(), serviceName)
			}
		case semconv.ServiceVersionKey:
			hasServiceVersion = true
			if attr.Value.AsString() != serviceVersion {
				t.Errorf("NewResource() service version = %q, want %q", attr.Value.AsString(), serviceVersion)
			}
		case semconv.HostNameKey:
			hasHostName = true
		}
	}

	if !hasServiceName {
		t.Error("NewResource() missing service name attribute")
	}
	if !hasServiceVersion {
		t.Error("NewResource() missing service version attribute")
	}
	if !hasHostName {
		t.Error("NewResource() missing host name attribute")
	}
}

func TestNewResourceWithEmptyValues(t *testing.T) {
	res := NewResource("", "")

	if res == nil {
		t.Fatal("NewResource() with empty values returned nil")
	}

	// Should still have attributes, even if empty
	attrs := res.Attributes()
	if len(attrs) == 0 {
		t.Error("NewResource() with empty values should still have attributes")
	}
}

func TestNewResourceHostname(t *testing.T) {
	res := NewResource("test", "1.0.0")

	// Get actual hostname
	expectedHostname, _ := os.Hostname()

	attrs := res.Attributes()
	for _, attr := range attrs {
		if attr.Key == semconv.HostNameKey {
			if attr.Value.AsString() != expectedHostname {
				t.Errorf("NewResource() hostname = %q, want %q", attr.Value.AsString(), expectedHostname)
			}
			return
		}
	}
	t.Error("NewResource() missing hostname attribute")
}

func TestNewResourceSchemaURL(t *testing.T) {
	res := NewResource("test", "1.0.0")

	schemaURL := res.SchemaURL()
	if schemaURL != semconv.SchemaURL {
		t.Errorf("NewResource() schema URL = %q, want %q", schemaURL, semconv.SchemaURL)
	}
}

// Tests for provider functions - these require OTLP endpoint
// Skip if not configured

func skipIfNoOTLPEndpoint(t *testing.T) {
	t.Helper()
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		t.Skip("skipping test: OTEL_EXPORTER_OTLP_ENDPOINT not set")
	}
}

func TestNewLoggerProvider(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()
	res := NewResource("test-service", "1.0.0")

	lp, err := NewLoggerProvider(ctx, res, true)
	if err != nil {
		t.Fatalf("NewLoggerProvider() error: %v", err)
	}

	if lp == nil {
		t.Fatal("NewLoggerProvider() returned nil provider")
	}

	// Clean up
	if err := lp.Shutdown(ctx); err != nil {
		t.Logf("LoggerProvider.Shutdown() error: %v", err)
	}
}

func TestNewLoggerProviderInsecure(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()
	res := NewResource("test-service", "1.0.0")

	// Test with insecure = false
	lp, err := NewLoggerProvider(ctx, res, false)
	if err != nil {
		// May fail without TLS configured
		t.Logf("NewLoggerProvider() with secure connection error: %v", err)
		return
	}

	if lp != nil {
		if err := lp.Shutdown(ctx); err != nil {
			t.Logf("LoggerProvider.Shutdown() error: %v", err)
		}
	}
}

func TestNewMeterProvider(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()
	res := NewResource("test-service", "1.0.0")

	mp, err := NewMeterProvider(ctx, res, true)
	if err != nil {
		t.Fatalf("NewMeterProvider() error: %v", err)
	}

	if mp == nil {
		t.Fatal("NewMeterProvider() returned nil provider")
	}

	// Clean up
	if err := mp.Shutdown(ctx); err != nil {
		t.Logf("MeterProvider.Shutdown() error: %v", err)
	}
}

func TestNewTracerProvider(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()
	res := NewResource("test-service", "1.0.0")

	tp, err := NewTracerProvider(ctx, res, true)
	if err != nil {
		t.Fatalf("NewTracerProvider() error: %v", err)
	}

	if tp == nil {
		t.Fatal("NewTracerProvider() returned nil provider")
	}

	// Clean up
	if err := tp.Shutdown(ctx); err != nil {
		t.Logf("TracerProvider.Shutdown() error: %v", err)
	}
}

// Test provider creation with nil resource
func TestNewLoggerProviderWithNilResource(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()

	lp, err := NewLoggerProvider(ctx, nil, true)
	if err != nil {
		t.Logf("NewLoggerProvider() with nil resource error: %v", err)
		return
	}

	if lp != nil {
		if err := lp.Shutdown(ctx); err != nil {
			t.Logf("LoggerProvider.Shutdown() error: %v", err)
		}
	}
}

func TestNewMeterProviderWithNilResource(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()

	mp, err := NewMeterProvider(ctx, nil, true)
	if err != nil {
		t.Logf("NewMeterProvider() with nil resource error: %v", err)
		return
	}

	if mp != nil {
		if err := mp.Shutdown(ctx); err != nil {
			t.Logf("MeterProvider.Shutdown() error: %v", err)
		}
	}
}

func TestNewTracerProviderWithNilResource(t *testing.T) {
	skipIfNoOTLPEndpoint(t)

	ctx := context.Background()

	tp, err := NewTracerProvider(ctx, nil, true)
	if err != nil {
		t.Logf("NewTracerProvider() with nil resource error: %v", err)
		return
	}

	if tp != nil {
		if err := tp.Shutdown(ctx); err != nil {
			t.Logf("TracerProvider.Shutdown() error: %v", err)
		}
	}
}

// Test with cancelled context
func TestNewLoggerProviderWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res := NewResource("test-service", "1.0.0")

	_, err := NewLoggerProvider(ctx, res, true)
	if err == nil {
		t.Log("NewLoggerProvider() with cancelled context did not return error (may succeed if connection not required)")
	}
}

func TestNewMeterProviderWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res := NewResource("test-service", "1.0.0")

	_, err := NewMeterProvider(ctx, res, true)
	if err == nil {
		t.Log("NewMeterProvider() with cancelled context did not return error (may succeed if connection not required)")
	}
}

func TestNewTracerProviderWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res := NewResource("test-service", "1.0.0")

	_, err := NewTracerProvider(ctx, res, true)
	if err == nil {
		t.Log("NewTracerProvider() with cancelled context did not return error (may succeed if connection not required)")
	}
}

// Test resource equality
func TestNewResourceEquality(t *testing.T) {
	res1 := NewResource("service1", "1.0.0")
	res2 := NewResource("service1", "1.0.0")
	res3 := NewResource("service2", "1.0.0")

	// Same service name and version should produce equivalent resources
	// (though not necessarily identical due to hostname)
	if res1.Equal(res3) {
		t.Error("Resources with different service names should not be equal")
	}

	// Check that both resources have same service name
	var res1Name, res2Name string
	for _, attr := range res1.Attributes() {
		if attr.Key == semconv.ServiceNameKey {
			res1Name = attr.Value.AsString()
		}
	}
	for _, attr := range res2.Attributes() {
		if attr.Key == semconv.ServiceNameKey {
			res2Name = attr.Value.AsString()
		}
	}

	if res1Name != res2Name {
		t.Errorf("Resources with same input should have same service name: %q != %q", res1Name, res2Name)
	}
}

// Benchmark tests
func BenchmarkNewResource(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewResource("benchmark-service", "1.0.0")
	}
}

func BenchmarkResourceAttributes(b *testing.B) {
	res := NewResource("benchmark-service", "1.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = res.Attributes()
	}
}

// Test that providers can be created without environment variables
// (they will fail to connect but should not panic)
func TestProviderCreationWithoutEndpoint(t *testing.T) {
	// Temporarily unset OTLP endpoint
	originalEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", originalEndpoint)
		}
	}()

	ctx := context.Background()
	res := NewResource("test-service", "1.0.0")

	// These should not panic, but may return errors
	_, err := NewLoggerProvider(ctx, res, true)
	if err != nil {
		t.Logf("NewLoggerProvider() without endpoint error: %v", err)
	}

	_, err = NewMeterProvider(ctx, res, true)
	if err != nil {
		t.Logf("NewMeterProvider() without endpoint error: %v", err)
	}

	_, err = NewTracerProvider(ctx, res, true)
	if err != nil {
		t.Logf("NewTracerProvider() without endpoint error: %v", err)
	}
}

// Test custom resource merging
func TestNewResourceMerge(t *testing.T) {
	res1 := NewResource("service1", "1.0.0")

	// Create another resource with additional attributes
	customRes, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNamespace("test-namespace"),
		),
	)

	// Merge resources
	merged, err := resource.Merge(res1, customRes)
	if err != nil {
		t.Fatalf("resource.Merge() error: %v", err)
	}

	if merged == nil {
		t.Fatal("resource.Merge() returned nil")
	}

	// Verify merged resource has attributes from both
	attrs := merged.Attributes()
	var hasServiceName, hasNamespace bool
	for _, attr := range attrs {
		if attr.Key == semconv.ServiceNameKey {
			hasServiceName = true
		}
		if attr.Key == semconv.ServiceNamespaceKey {
			hasNamespace = true
		}
	}

	if !hasServiceName {
		t.Error("Merged resource missing service name")
	}
	if !hasNamespace {
		t.Error("Merged resource missing namespace")
	}
}
