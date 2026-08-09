package telemetry

import (
	"context"
	"testing"
	"time"
)

// TestNewGCPGRPCConn exercises NewGCPGRPCConn's contract: it must return
// promptly (bounded by the context passed in) with either a usable
// connection or an error, never panicking or blocking past the deadline.
// Application Default Credentials are not guaranteed to be available in the
// test environment, so this test does not assert success — only that the
// function behaves well in both outcomes.
func TestNewGCPGRPCConn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := NewGCPGRPCConn(ctx)
	if err != nil {
		t.Logf("NewGCPGRPCConn returned an error (expected without ADC credentials): %v", err)
		return
	}

	if conn == nil {
		t.Fatalf("expected a non-nil connection when no error is returned")
	}
	defer func() { _ = conn.Close() }()
}
