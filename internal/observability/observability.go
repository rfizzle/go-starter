// Package observability wires the application's OpenTelemetry signals to an
// OTLP/gRPC collector.
//
// The generated ogen server and client already create spans and record metrics
// against OpenTelemetry's *global* providers (see internal/api/v1). Out of the
// box those globals are no-op implementations, so every span and metric the
// generated code produces is silently discarded. Init installs real SDK
// providers backed by OTLP exporters and registers them as the globals, so the
// already-instrumented code starts exporting without any further wiring.
//
// Init is a no-op when observability is disabled in config: the no-op globals
// stay in place, the exporters are never dialled, and the returned shutdown
// function does nothing. This keeps the exporter dependency inert until an
// operator opts in, so a fresh clone runs with zero telemetry overhead.
//
// This owns logger-adjacent construction only for traces and meters; the zap
// logger is still built in pkg/config. Consolidating all three here is the
// longer-term pattern (see the hardening review), but the logger is
// intentionally left where it is for now to keep this change focused.
package observability

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	// Match the semconv version resource.Default() is built from (see the OTel
	// SDK's resource package) so resource.Merge below sees a single schema URL
	// and does not report a conflict.
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/rfizzle/go-starter/pkg/config"
)

// Init installs OpenTelemetry SDK providers for traces and metrics, wired to an
// OTLP/gRPC collector, and registers them as the process-wide global providers
// so the already-instrumented ogen code begins exporting. It also installs a
// W3C trace-context + baggage propagator so trace context survives across
// service boundaries.
//
// The returned shutdown function flushes and closes both providers; it MUST be
// called during graceful shutdown so buffered spans and metrics are exported
// rather than dropped. It is always safe to call (and safe to call on the
// disabled path, where it is a no-op).
//
// When cfg.Observability.Enabled is false, Init returns immediately with a
// no-op shutdown and leaves the global providers untouched.
func Init(ctx context.Context, cfg *config.Application) (shutdown func(context.Context) error, err error) {
	if !cfg.Observability.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	// Collect each installed component's shutdown so a partial failure during
	// setup can still tear down whatever already started, and so the caller gets
	// a single function that closes everything in reverse order.
	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var errs error
		// Close in reverse registration order.
		for i := len(shutdownFuncs) - 1; i >= 0; i-- {
			if sErr := shutdownFuncs[i](ctx); sErr != nil {
				errs = errors.Join(errs, sErr)
			}
		}
		shutdownFuncs = nil
		return errs
	}

	// On any setup error, tear down whatever we already installed so we never
	// leak a live exporter connection back to the caller.
	handleErr := func(inErr error) (func(context.Context) error, error) {
		return shutdown, errors.Join(inErr, shutdown(ctx))
	}

	res, err := newResource(cfg)
	if err != nil {
		return handleErr(fmt.Errorf("building resource: %w", err))
	}

	// Propagate W3C trace context and baggage so spans join up across services.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracerProvider, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return handleErr(fmt.Errorf("building tracer provider: %w", err))
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, cfg, res)
	if err != nil {
		return handleErr(fmt.Errorf("building meter provider: %w", err))
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return shutdown, nil
}

// newResource describes this service to the collector. The attributes here
// (service name, version, deployment environment) become the resource labels
// every exported span and metric is tagged with, so telemetry from this binary
// is distinguishable from every other service in the same collector.
func newResource(cfg *config.Application) (*resource.Resource, error) {
	serviceName := cfg.Name
	if serviceName == "" {
		serviceName = "go-starter"
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
	}
	if cfg.Build.Version != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(cfg.Build.Version))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentNameKey.String(cfg.Environment))
	}

	// Tag our attributes with the semconv schema URL they came from, then merge
	// them over resource.Default() so the SDK's own attributes (telemetry.sdk.*)
	// are preserved rather than replaced. Both resources share the same schema
	// URL, so Merge does not report a conflict.
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}

// newTracerProvider builds a batching tracer provider exporting over OTLP/gRPC.
// A parent-based sampler wraps the configured ratio so an upstream sampling
// decision is always honoured, keeping distributed traces complete.
func newTracerProvider(ctx context.Context, cfg *config.Application, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, traceExporterOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Observability.SampleRatio))

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	), nil
}

// newMeterProvider builds a periodic-reading meter provider exporting over
// OTLP/gRPC.
func newMeterProvider(ctx context.Context, cfg *config.Application, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx, metricExporterOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP metric exporter: %w", err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	), nil
}

// traceExporterOptions maps config onto the OTLP/gRPC trace exporter options.
// An empty endpoint is left unset so the exporter falls back to the OTel SDK's
// own environment-variable resolution and then its localhost default.
func traceExporterOptions(cfg *config.Application) []otlptracegrpc.Option {
	opts := []otlptracegrpc.Option{}
	if cfg.Observability.Endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(cfg.Observability.Endpoint))
	}
	if cfg.Observability.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return opts
}

// metricExporterOptions mirrors traceExporterOptions for the metric exporter.
func metricExporterOptions(cfg *config.Application) []otlpmetricgrpc.Option {
	opts := []otlpmetricgrpc.Option{}
	if cfg.Observability.Endpoint != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(cfg.Observability.Endpoint))
	}
	if cfg.Observability.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	return opts
}
