package telemetry

import (
	"log/slog"

	"go.opentelemetry.io/otel/metric"
)

// NewInt64Counter creates an Int64Counter on meter, logging (but not
// panicking) on error.
func NewInt64Counter(meter metric.Meter, name, description string) metric.Int64Counter {
	c, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		slog.Error("failed to create metric counter", slog.String("name", name), slog.String("error", err.Error()))
	}
	return c
}

// NewInt64Histogram creates an Int64Histogram on meter, logging (but not
// panicking) on error. Additional options (for example,
// metric.WithUnit("By")) may be supplied.
func NewInt64Histogram(meter metric.Meter, name, description string, opts ...metric.Int64HistogramOption) metric.Int64Histogram {
	allOpts := append([]metric.Int64HistogramOption{metric.WithDescription(description)}, opts...)
	h, err := meter.Int64Histogram(name, allOpts...)
	if err != nil {
		slog.Error("failed to create metric histogram", slog.String("name", name), slog.String("error", err.Error()))
	}
	return h
}
