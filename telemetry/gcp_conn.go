package telemetry

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"google.golang.org/grpc"
)

// gcpOTLPEndpoint is GCP's managed OTel collector endpoint.
const gcpOTLPEndpoint = "telemetry.googleapis.com:443"

// NewGCPGRPCConn creates a single gRPC client connection to GCP's managed
// OTel collector endpoint, authenticated via Application Default Credentials
// (ADC). The connection is intended to be shared by all three OTLP exporters
// (traces, logs, metrics) via WithGRPCConn, so that only one set of TLS
// handshakes and token-refresh goroutines is needed.
//
// ADC credential resolution order:
//  1. GOOGLE_APPLICATION_CREDENTIALS env var — path to a service account key
//     JSON file; use this for non-GCE environments (local dev, non-GCP servers)
//  2. gcloud auth application-default login — credentials on the local machine
//  3. GCE / Cloud Run metadata server — zero-config for VM-hosted workloads;
//     tokens are fetched and refreshed automatically via the metadata server
//
// Token refresh is handled transparently by the google.golang.org/api
// transport layer — no manual management of Bearer tokens is required.
//
// Callers that need to bound how long ADC resolution may block (for example,
// so a developer running locally without credentials does not stall
// indefinitely on an unreachable metadata server) should pass a ctx with a
// deadline.
func NewGCPGRPCConn(ctx context.Context) (*grpc.ClientConn, error) {
	return transport.DialGRPC(ctx,
		option.WithEndpoint(gcpOTLPEndpoint),
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"),
	)
}
