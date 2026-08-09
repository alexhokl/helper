package telemetry

import (
	"testing"

	"go.opentelemetry.io/otel"
)

func TestNewInt64Counter(t *testing.T) {
	meter := otel.Meter("test-meter")
	c := NewInt64Counter(meter, "test.counter", "a test counter")
	if c == nil {
		t.Fatalf("expected a non-nil counter")
	}
}

func TestNewInt64Histogram(t *testing.T) {
	meter := otel.Meter("test-meter")
	h := NewInt64Histogram(meter, "test.histogram", "a test histogram")
	if h == nil {
		t.Fatalf("expected a non-nil histogram")
	}
}
