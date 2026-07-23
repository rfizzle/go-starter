package health

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAggregator_CheckBoundsSlowDependency(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const budget = 25 * time.Millisecond
		aggregator, err := New(budget, zap.NewNop(), Check{
			Name: "postgres",
			Run: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
		})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		started := time.Now()
		err = aggregator.Check(context.Background())
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Check() error = %v, want deadline exceeded", err)
		}
		if elapsed := time.Since(started); elapsed != budget {
			t.Errorf("Check() elapsed = %v, want %v budget", elapsed, budget)
		}
	})
}

func TestAggregator_LogsOnlyReadinessTransitions(t *testing.T) {
	var output bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&output), zap.InfoLevel)
	var unavailable atomic.Bool
	secretError := errors.New("postgres://user:password@database/private: connection refused")

	aggregator, err := New(time.Second, zap.New(core), Check{
		Name: "postgres",
		Run: func(context.Context) error {
			if unavailable.Load() {
				return secretError
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := aggregator.Check(t.Context()); err != nil {
		t.Fatalf("healthy Check() error = %v", err)
	}
	unavailable.Store(true)
	for range 2 {
		if err := aggregator.Check(t.Context()); !errors.Is(err, secretError) {
			t.Fatalf("unhealthy Check() error = %v, want wrapped dependency error", err)
		}
	}
	unavailable.Store(false)
	if err := aggregator.Check(t.Context()); err != nil {
		t.Fatalf("recovered Check() error = %v", err)
	}

	logs := output.String()
	if count := strings.Count(logs, `"msg":"readiness dependency state changed"`); count != 3 {
		t.Fatalf("transition log count = %d, want 3; logs: %s", count, logs)
	}
	if count := strings.Count(logs, `"failed_check":"postgres"`); count != 1 {
		t.Errorf("failed_check log count = %d, want 1; logs: %s", count, logs)
	}
	if strings.Contains(logs, "password") {
		t.Errorf("transition log exposed dependency error: %s", logs)
	}
}

func TestNewRejectsInvalidChecks(t *testing.T) {
	valid := Check{Name: "postgres", Run: func(context.Context) error { return nil }}
	tests := []struct {
		name    string
		timeout time.Duration
		logger  *zap.Logger
		checks  []Check
	}{
		{name: "unbounded timeout", logger: zap.NewNop(), checks: []Check{valid}},
		{name: "nil logger", timeout: time.Second, checks: []Check{valid}},
		{name: "no checks", timeout: time.Second, logger: zap.NewNop()},
		{
			name:    "duplicate name",
			timeout: time.Second,
			logger:  zap.NewNop(),
			checks:  []Check{valid, valid},
		},
		{
			name:    "missing function",
			timeout: time.Second,
			logger:  zap.NewNop(),
			checks:  []Check{{Name: "postgres"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.timeout, tt.logger, tt.checks...); err == nil {
				t.Fatal("New() error = nil, want invalid configuration error")
			}
		})
	}
}
