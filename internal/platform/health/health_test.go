package health

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

func TestAggregator_CheckBoundsSlowDependency(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const budget = 25 * time.Millisecond
		aggregator, err := New(budget, Check{
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

func TestNewRejectsInvalidChecks(t *testing.T) {
	valid := Check{Name: "postgres", Run: func(context.Context) error { return nil }}
	tests := []struct {
		name    string
		timeout time.Duration
		checks  []Check
	}{
		{name: "unbounded timeout", checks: []Check{valid}},
		{name: "no checks", timeout: time.Second},
		{name: "missing name", timeout: time.Second, checks: []Check{{Name: " ", Run: valid.Run}}},
		{name: "missing function", timeout: time.Second, checks: []Check{{Name: "postgres"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.timeout, tt.checks...); err == nil {
				t.Fatal("New() error = nil, want invalid configuration error")
			}
		})
	}
}
