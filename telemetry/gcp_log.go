package telemetry

import (
	"log/slog"
	"os"
)

// Field names recognised by Google Cloud Logging when a structured JSON
// payload is written to stdout. Emitting slog's default keys (level, msg,
// time) instead would cause Cloud Logging to treat the record as an opaque
// text payload, so every entry would be ingested at default severity and
// would not be filterable by log level.
//
// See https://cloud.google.com/logging/docs/structured-logging
const (
	GCPSeverityKey  = "severity"
	GCPMessageKey   = "message"
	GCPTimestampKey = "timestamp"
)

// NewGCPConsoleHandler returns a console log handler suitable for fanning
// out alongside the OpenTelemetry bridge, so that application logs remain
// visible even when telemetry export is failing.
//
// When running on GCP (see IsRunningOnGCP), a JSON handler writing to stdout
// is returned so that the container logging driver forwards well-formed
// structured entries to Cloud Logging. Elsewhere, fallback is returned
// unchanged, leaving local output (for example, plain text on stderr)
// untouched.
func NewGCPConsoleHandler(fallback slog.Handler) slog.Handler {
	if !IsRunningOnGCP() {
		return fallback
	}

	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: ReplaceAttrForGCP,
	})
}

// ReplaceAttrForGCP renames slog's built-in top-level attributes to the
// field names Cloud Logging recognises. Attributes nested inside a group are
// left alone, as the special field names are only meaningful at the root of
// the payload.
func ReplaceAttrForGCP(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.LevelKey:
		attr.Key = GCPSeverityKey
	case slog.MessageKey:
		attr.Key = GCPMessageKey
	case slog.TimeKey:
		attr.Key = GCPTimestampKey
	}

	return attr
}

// IsRunningOnGCP reports whether the process appears to be running on
// Google Cloud, based on the environment variables set by Cloud Run and
// commonly used by GCP deployment configurations.
func IsRunningOnGCP() bool {
	for _, name := range []string{
		"GOOGLE_CLOUD_PROJECT",
		"K_SERVICE", // set by Cloud Run
	} {
		if os.Getenv(name) != "" {
			return true
		}
	}
	return false
}
