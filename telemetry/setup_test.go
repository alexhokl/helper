package telemetry

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIsOTLPConfigured(t *testing.T) {
	envNames := []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	}

	tests := []struct {
		name     string
		set      string
		expected bool
	}{
		{"none set", "", false},
		{"general endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT", true},
		{"traces endpoint", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", true},
		{"metrics endpoint", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", true},
		{"logs endpoint", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, name := range envNames {
				t.Setenv(name, "")
			}
			if test.set != "" {
				t.Setenv(test.set, "localhost:4317")
			}
			if result := IsOTLPConfigured(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestIsSDKDisabled(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"unset", "", false},
		{"true", "true", true},
		{"True", "True", true},
		{"one", "1", true},
		{"padded", "  true  ", true},
		{"false", "false", false},
		{"nonsense", "yes-please", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("OTEL_SDK_DISABLED", test.value)
			if result := IsSDKDisabled(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

// TestSetupWithoutEndpoint asserts that a server started without a collector
// keeps its default slog handler, so log records are not silently discarded.
func TestSetupWithoutEndpoint(t *testing.T) {
	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		t.Setenv(name, "")
	}

	before := slog.Default()

	shutdown, err := Setup(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function")
	}
	if slog.Default() != before {
		t.Errorf("expected the default slog logger to be left untouched")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("expected no error from shutdown, got %v", err)
	}
}

// TestSetupDisabledViaCheck asserts the WithDisabledCheck option is honoured,
// leaving the default logger untouched, mirroring OTEL_SDK_DISABLED-based
// gating.
func TestSetupDisabledViaCheck(t *testing.T) {
	before := slog.Default()

	shutdown, err := Setup(
		context.Background(),
		WithDisabledCheck(func() bool { return true }),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function")
	}
	if slog.Default() != before {
		t.Errorf("expected the default logger to be left untouched when disabled")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("expected no error from shutdown, got %v", err)
	}
}

// TestSetupMissingServiceNameReturnsError asserts that Setup requires
// OTEL_SERVICE_NAME to be set whenever telemetry is enabled, and that the
// failure leaves the default logger untouched, just like any other Setup
// failure.
func TestSetupMissingServiceNameReturnsError(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")

	before := slog.Default()

	shutdown, err := Setup(
		context.Background(),
		WithDisabledCheck(func() bool { return false }),
	)
	if err == nil {
		t.Fatalf("expected an error when OTEL_SERVICE_NAME is unset")
	}
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function even on failure")
	}
	if slog.Default() != before {
		t.Errorf("expected the default logger to be left untouched on failure")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("expected no error from shutdown, got %v", err)
	}
}

// TestSetupClosesGRPCConnEvenWhenServiceNameMissing asserts that a
// caller-supplied gRPC connection (via WithGRPCConn) is always closed by the
// returned shutdown function, even when Setup fails before constructing any
// exporter — such as when OTEL_SERVICE_NAME is missing.
func TestSetupClosesGRPCConnEvenWhenServiceNameMissing(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")

	conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create test grpc conn: %v", err)
	}

	shutdown, err := Setup(
		context.Background(),
		WithGRPCConn(conn),
		WithDisabledCheck(func() bool { return false }),
	)
	if err == nil {
		t.Fatalf("expected an error when OTEL_SERVICE_NAME is unset")
	}

	if shutdownErr := shutdown(context.Background()); shutdownErr != nil {
		t.Errorf("expected no error from shutdown, got %v", shutdownErr)
	}

	if state := conn.GetState(); state != connectivity.Shutdown {
		t.Errorf("expected the supplied grpc conn to be closed (shutdown), got state %v", state)
	}
}

// TestSetupMissingEndpointReturnsError asserts that Setup requires
// OTEL_EXPORTER_OTLP_ENDPOINT to be set whenever telemetry is enabled and no
// WithGRPCConn was supplied, and that the error suggests a sensible default
// for a local collector.
func TestSetupMissingEndpointReturnsError(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "endpoint-test")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	before := slog.Default()

	shutdown, err := Setup(
		context.Background(),
		WithDisabledCheck(func() bool { return false }),
	)
	if err == nil {
		t.Fatalf("expected an error when OTEL_EXPORTER_OTLP_ENDPOINT is unset")
	}
	if !strings.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Errorf("expected error to mention OTEL_EXPORTER_OTLP_ENDPOINT, got: %v", err)
	}
	if !strings.Contains(err.Error(), "http://localhost:4317") {
		t.Errorf("expected error to suggest http://localhost:4317, got: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function even on failure")
	}
	if slog.Default() != before {
		t.Errorf("expected the default logger to be left untouched on failure")
	}
	if shutdownErr := shutdown(context.Background()); shutdownErr != nil {
		t.Errorf("expected no error from shutdown, got %v", shutdownErr)
	}
}

// TestSetupEndpointCheckSkippedWithGRPCConn asserts that the
// OTEL_EXPORTER_OTLP_ENDPOINT requirement does not apply when the caller
// supplies its own connection via WithGRPCConn, since the exporters use that
// connection directly and never consult the endpoint variable.
func TestSetupEndpointCheckSkippedWithGRPCConn(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "endpoint-skip-test")
	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		t.Setenv(name, "")
	}

	// Setup mutates global provider/logger state on success; save and
	// restore it so this test cannot leak into others.
	originalTracerProvider := otel.GetTracerProvider()
	originalMeterProvider := otel.GetMeterProvider()
	originalLogger := slog.Default()
	t.Cleanup(func() {
		otel.SetTracerProvider(originalTracerProvider)
		otel.SetMeterProvider(originalMeterProvider)
		slog.SetDefault(originalLogger)
	})

	conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create test grpc conn: %v", err)
	}

	shutdown, err := Setup(
		context.Background(),
		WithGRPCConn(conn),
		WithDisabledCheck(func() bool { return false }),
	)
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function")
	}
	t.Cleanup(func() {
		// Bounded so this test doesn't pay the SDK's full metric-export
		// timeout (10s) trying to flush against the fake, unreachable
		// "passthrough" target used here.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		_ = shutdown(shutdownCtx)
	})

	if err != nil && strings.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Errorf("expected the endpoint check to be skipped when WithGRPCConn is supplied, got error: %v", err)
	}
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{"unset", "", ""},
		{"set", "custom-service", "custom-service"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("OTEL_SERVICE_NAME", test.envValue)
			if result := ServiceName(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestBuildResource(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")

	res, err := BuildResource(context.Background(), "resource-test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var found bool
	for _, attr := range res.Attributes() {
		if attr.Key == semconv.ServiceNameKey && attr.Value.AsString() == "resource-test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected resource to carry service.name of resource-test")
	}
}

func TestBuildResourceWithExtraAttributes(t *testing.T) {
	res, err := BuildResource(context.Background(), "resource-test", semconv.CloudProviderGCP)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var found bool
	for _, attr := range res.Attributes() {
		if attr.Key == semconv.CloudProviderKey && attr.Value.AsString() == semconv.CloudProviderGCP.Value.AsString() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected resource to carry the supplied extra attribute")
	}
}
