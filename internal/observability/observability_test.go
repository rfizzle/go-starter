package observability

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/rfizzle/go-starter/pkg/config"
)

// TestInitDisabledIsNoop verifies that when observability is disabled Init does
// not dial an exporter and returns a shutdown that is safe to call. This is the
// default path for a fresh clone, so it must never fail or block.
func TestInitDisabledIsNoop(t *testing.T) {
	t.Parallel()

	cfg := &config.Application{}
	cfg.Observability.Enabled = false

	shutdown, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init() error = %v, want nil on disabled path", err)
	}
	if shutdown == nil {
		t.Fatal("Init() returned nil shutdown; want a callable no-op")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown() error = %v, want nil on disabled path", err)
	}
}

// TestTraceExporterOptions checks the config-to-option mapping without opening a
// connection: an empty endpoint contributes no endpoint option (so the SDK's
// env/default resolution wins) and Insecure is only appended when set.
func TestTraceExporterOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		insecure bool
		wantLen  int
	}{
		{name: "empty endpoint, secure", endpoint: "", insecure: false, wantLen: 0},
		{name: "endpoint only", endpoint: "localhost:4317", insecure: false, wantLen: 1},
		{name: "insecure only", endpoint: "", insecure: true, wantLen: 1},
		{name: "endpoint and insecure", endpoint: "collector:4317", insecure: true, wantLen: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Application{}
			cfg.Observability.Endpoint = tc.endpoint
			cfg.Observability.Insecure = tc.insecure

			if got := len(traceExporterOptions(cfg)); got != tc.wantLen {
				t.Errorf("traceExporterOptions() len = %d, want %d", got, tc.wantLen)
			}
			if got := len(metricExporterOptions(cfg)); got != tc.wantLen {
				t.Errorf("metricExporterOptions() len = %d, want %d", got, tc.wantLen)
			}
		})
	}
}

// TestNewResourceAttributes confirms the service is described to the collector
// with a name, version, and environment, and that an empty name falls back to a
// stable default rather than an unlabelled resource.
func TestNewResourceAttributes(t *testing.T) {
	t.Parallel()

	t.Run("populated", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Application{Name: "svc", Environment: "staging"}
		cfg.Build.Version = "1.2.3"

		res, err := newResource(cfg)
		if err != nil {
			t.Fatalf("newResource() error = %v", err)
		}

		want := map[string]string{
			string(semconv.ServiceNameKey):               "svc",
			string(semconv.ServiceVersionKey):            "1.2.3",
			string(semconv.DeploymentEnvironmentNameKey): "staging",
		}
		got := attrMap(res.Attributes())
		for k, v := range want {
			if got[k] != v {
				t.Errorf("resource attribute %q = %q, want %q", k, got[k], v)
			}
		}
		// resource.Merge with resource.Default() must preserve the SDK's own
		// attributes rather than replacing them.
		if got[string(semconv.TelemetrySDKLanguageKey)] == "" {
			t.Error("resource missing telemetry.sdk.language; SDK defaults were not merged")
		}
	})

	t.Run("defaults service name", func(t *testing.T) {
		t.Parallel()

		res, err := newResource(&config.Application{})
		if err != nil {
			t.Fatalf("newResource() error = %v", err)
		}

		got := attrMap(res.Attributes())
		if got[string(semconv.ServiceNameKey)] != "go-starter" {
			t.Errorf("service.name = %q, want %q (default)", got[string(semconv.ServiceNameKey)], "go-starter")
		}
		if _, ok := got[string(semconv.ServiceVersionKey)]; ok {
			t.Error("service.version set with empty build version; want omitted")
		}
		if _, ok := got[string(semconv.DeploymentEnvironmentNameKey)]; ok {
			t.Error("deployment.environment.name set with empty environment; want omitted")
		}
	})
}

// TestInitEnabledInstallsProviders verifies the enabled path installs real
// global providers and a working shutdown. The OTLP/gRPC exporter uses
// grpc.NewClient, which connects lazily, so Init succeeds with no collector
// listening — this exercises the wiring without a live backend. It mutates
// process-global OTel state, so it must not run in parallel and restores the
// globals afterwards.
func TestInitEnabledInstallsProviders(t *testing.T) {
	origTracer := otel.GetTracerProvider()
	origMeter := otel.GetMeterProvider()
	origProp := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(origTracer)
		otel.SetMeterProvider(origMeter)
		otel.SetTextMapPropagator(origProp)
	})

	cfg := &config.Application{Name: "svc"}
	cfg.Observability.Enabled = true
	cfg.Observability.Endpoint = "localhost:4317"
	cfg.Observability.Insecure = true
	cfg.Observability.SampleRatio = 1.0

	shutdown, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init() error = %v, want nil on enabled path", err)
	}

	// The composite propagator must carry both W3C trace context and baggage so
	// distributed traces and baggage survive across service boundaries.
	fields := otel.GetTextMapPropagator().Fields()
	assertContains(t, fields, "traceparent")
	assertContains(t, fields, "baggage")

	// The installed providers must be the real SDK ones, not the no-op globals.
	if otel.GetTracerProvider() == origTracer {
		t.Error("tracer provider was not replaced with an SDK provider")
	}
	if otel.GetMeterProvider() == origMeter {
		t.Error("meter provider was not replaced with an SDK provider")
	}

	// shutdown must return within its deadline. It may report an error because no
	// collector is listening (the final flush can't reach one) — that is the same
	// error app.go logs as a warning, not a test failure. We only assert that the
	// call completes rather than hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = shutdown(ctx)
}

// assertContains fails the test if want is not present in got.
func assertContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, g := range got {
		if g == want {
			return
		}
	}
	t.Errorf("propagator fields %v missing %q", got, want)
}

// attrMap flattens a resource's attribute set into a string map for assertions.
func attrMap(kvs []attribute.KeyValue) map[string]string {
	out := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		out[string(kv.Key)] = kv.Value.AsString()
	}
	return out
}
