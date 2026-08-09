// Package telemetry provides shared OpenTelemetry setup, span helpers, gRPC
// interceptors, metric constructors, and GCP-specific logging helpers used
// across services.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

// setupConfig holds the options accumulated by the functional Option values
// passed to Setup.
type setupConfig struct {
	grpcConn      *grpc.ClientConn
	resourceAttrs []attribute.KeyValue
	disabled      func() bool
	logSink       func(lp *sdklog.LoggerProvider) slog.Handler
}

// Option configures Setup.
type Option func(*setupConfig)

// WithGRPCConn supplies a pre-established gRPC connection (for example, one
// authenticated via Application Default Credentials to a cloud vendor's
// managed OTLP endpoint) to be shared by all three OTLP exporters. When not
// supplied, each exporter dials the endpoint configured through the standard
// OTEL_EXPORTER_OTLP_* environment variables. The connection is closed as
// part of the returned shutdown function.
func WithGRPCConn(conn *grpc.ClientConn) Option {
	return func(c *setupConfig) {
		c.grpcConn = conn
	}
}

// WithResourceAttributes appends additional attributes to the OTel resource
// describing this service (for example, cloud provider/project/region).
func WithResourceAttributes(attrs ...attribute.KeyValue) Option {
	return func(c *setupConfig) {
		c.resourceAttrs = append(c.resourceAttrs, attrs...)
	}
}

// WithDisabledCheck overrides the function used to decide whether telemetry
// should be disabled entirely. The default is IsOTLPConfigured negated
// (telemetry is disabled when no OTLP endpoint is configured).
func WithDisabledCheck(fn func() bool) Option {
	return func(c *setupConfig) {
		c.disabled = fn
	}
}

// WithLogFanout replaces the default log sink (the otelslog bridge alone)
// with the return value of fn, which receives the constructed
// LoggerProvider. This allows callers to fan the log record out to
// additional sinks (for example, a console handler) so that a broken OTLP
// exporter cannot silence application logs. See NewGCPConsoleHandler.
func WithLogFanout(fn func(lp *sdklog.LoggerProvider) slog.Handler) Option {
	return func(c *setupConfig) {
		c.logSink = fn
	}
}

// Setup initialises the global OpenTelemetry TracerProvider, MeterProvider,
// and LoggerProvider, exporting all three signals over OTLP/gRPC.
//
// Traces: a span is started for every incoming gRPC call when the otelgrpc
// stats handler is registered by the caller (grpc.StatsHandler(otelgrpc.NewServerHandler())).
//
// Metrics: the MeterProvider is registered globally so instrumentation such
// as otelgrpc emits standard metrics automatically. Metrics are pushed at the
// SDK default interval (60s), overridable via OTEL_METRIC_EXPORT_INTERVAL.
//
// Logs: the default slog logger is replaced with a handler backed by the
// LoggerProvider (optionally fanned out to additional sinks via
// WithLogFanout) so that slog.InfoContext/WarnContext/ErrorContext calls emit
// structured OTLP log records carrying the active trace_id and span_id.
//
// The OTEL_SERVICE_NAME environment variable is required whenever telemetry
// is enabled; Setup returns an error if it is unset. This ensures every
// exported signal carries an accurate service.name resource attribute rather
// than falling back to a hardcoded default that could be forgotten in a new
// deployment environment.
//
// The OTEL_EXPORTER_OTLP_ENDPOINT environment variable is likewise required
// whenever telemetry is enabled and no WithGRPCConn was supplied; Setup
// returns an error if it is unset in that case. When WithGRPCConn is
// supplied (for example, a connection to a cloud vendor's managed OTLP
// endpoint authenticated out-of-band via ADC), the exporters use that
// connection directly and this variable is not consulted, so the check is
// skipped.
//
// When telemetry is disabled (see WithDisabledCheck, IsOTLPConfigured, and
// IsSDKDisabled), the default slog handler is left untouched entirely, so a
// server started without a collector still logs to stderr rather than
// silently discarding every record. If telemetry is enabled but a component
// cannot be constructed (including a missing OTEL_SERVICE_NAME or
// OTEL_EXPORTER_OTLP_ENDPOINT), an error is returned rather than swallowed.
//
// The returned shutdown function must be deferred by the caller; it flushes
// and shuts down all constructed providers (and closes any connection
// supplied via WithGRPCConn). It is always non-nil, even when an error is
// returned.
func Setup(ctx context.Context, opts ...Option) (shutdown func(context.Context) error, err error) {
	cfg := &setupConfig{
		disabled: func() bool { return !IsOTLPConfigured() },
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls each registered cleanup function and joins their errors.
	shutdown = func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			errs = append(errs, fn(ctx))
		}
		return errors.Join(errs...)
	}

	// Registered immediately so that a caller-supplied connection is always
	// closed via the returned shutdown function, regardless of where Setup
	// fails below (missing service name, disabled telemetry, or a
	// construction error).
	if cfg.grpcConn != nil {
		conn := cfg.grpcConn
		shutdownFuncs = append(shutdownFuncs, func(_ context.Context) error {
			return conn.Close()
		})
	}

	if cfg.disabled != nil && cfg.disabled() {
		slog.Info("OpenTelemetry is disabled")
		return shutdown, nil
	}

	name := ServiceName()
	if name == "" {
		return shutdown, fmt.Errorf("OTEL_SERVICE_NAME environment variable must be set")
	}

	// When no pre-established connection was supplied via WithGRPCConn, each
	// exporter dials the endpoint configured through OTEL_EXPORTER_OTLP_ENDPOINT,
	// so it must be set. Callers that supply their own connection (for
	// example, one authenticated via ADC to a cloud vendor's managed OTLP
	// endpoint) do not consult this variable at all, so the check does not
	// apply to them.
	if cfg.grpcConn == nil && os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return shutdown, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT environment variable must be set (try http://localhost:4317 if unsure)")
	}

	res, err := BuildResource(ctx, name, cfg.resourceAttrs...)
	if err != nil {
		return shutdown, fmt.Errorf("failed to build OpenTelemetry resource: %w", err)
	}

	// ── Traces ──────────────────────────────────────────────────────────────

	traceOpts := []otlptracegrpc.Option{}
	if cfg.grpcConn != nil {
		traceOpts = append(traceOpts, otlptracegrpc.WithGRPCConn(cfg.grpcConn))
	}
	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// ── Logs ────────────────────────────────────────────────────────────────

	logOpts := []otlploggrpc.Option{}
	if cfg.grpcConn != nil {
		logOpts = append(logOpts, otlploggrpc.WithGRPCConn(cfg.grpcConn))
	}
	logExporter, err := otlploggrpc.New(ctx, logOpts...)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, lp.Shutdown)

	// ── Metrics ─────────────────────────────────────────────────────────────

	metricOpts := []otlpmetricgrpc.Option{}
	if cfg.grpcConn != nil {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithGRPCConn(cfg.grpcConn))
	}
	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	// Route OTel SDK internal errors (including failed OTLP exports due to
	// auth rejection, TLS errors, or timeouts) to stderr so they are visible
	// rather than silently discarded.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		fmt.Fprintf(os.Stderr, "OpenTelemetry error: %v\n", err)
	}))

	// ── Default logger ─────────────────────────────────────────────────────
	//
	// Every provider has now been constructed, so it is safe to take over the
	// default logger. A setup failure earlier always leaves logging untouched.
	var handler slog.Handler
	if cfg.logSink != nil {
		handler = cfg.logSink(lp)
	} else {
		handler = otelslog.NewHandler(name, otelslog.WithLoggerProvider(lp))
	}
	slog.SetDefault(slog.New(handler))

	return shutdown, nil
}

// IsOTLPConfigured reports whether an OTLP endpoint has been configured
// through the standard environment variables, either globally or for any
// individual signal.
func IsOTLPConfigured() bool {
	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		if os.Getenv(name) != "" {
			return true
		}
	}
	return false
}

// ServiceName returns the service name from the OTEL_SERVICE_NAME
// environment variable. Setup requires this variable to be set whenever
// telemetry is enabled, so by the time any Setup-supplied callback (such as
// WithLogFanout) runs, it is guaranteed to be non-empty.
func ServiceName() string {
	return os.Getenv("OTEL_SERVICE_NAME")
}

// BuildResource builds an OTel resource describing this service, merging
// process, OS, and OTEL_RESOURCE_ATTRIBUTES-derived attributes with the given
// service name and any extra attributes supplied by the caller.
func BuildResource(ctx context.Context, serviceName string, extra ...attribute.KeyValue) (*resource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	attrs := append([]attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceInstanceID(hostname),
	}, extra...)

	return resource.New(
		ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithFromEnv(), // picks up OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME
		resource.WithAttributes(attrs...),
	)
}

// IsSDKDisabled reports whether telemetry has been switched off through the
// standard OTEL_SDK_DISABLED environment variable.
func IsSDKDisabled() bool {
	disabled, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")))
	if err != nil {
		return false
	}
	return disabled
}
