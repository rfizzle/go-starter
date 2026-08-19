// Package health evaluates the technical dependencies required for the
// application to serve traffic.
package health

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Check is a named readiness dependency. Name must be a stable, non-secret
// identifier, and Run must honor context cancellation.
type Check struct {
	Name string
	Run  func(context.Context) error
}

// Aggregator evaluates all required checks within one shared time budget.
// It is stateless between evaluations and safe for concurrent use.
type Aggregator struct {
	checks  []Check
	timeout time.Duration
}

// New constructs an Aggregator. Timeout must be positive so every evaluation
// is bounded, even when the caller has no deadline.
func New(timeout time.Duration, checks ...Check) (*Aggregator, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("health check timeout must be positive")
	}
	if len(checks) == 0 {
		return nil, fmt.Errorf("at least one health check is required")
	}

	configured := make([]Check, len(checks))
	for i, check := range checks {
		check.Name = strings.TrimSpace(check.Name)
		if check.Name == "" {
			return nil, fmt.Errorf("health check %d name is required", i)
		}
		if check.Run == nil {
			return nil, fmt.Errorf("health check %q function is required", check.Name)
		}
		configured[i] = check
	}

	return &Aggregator{
		checks:  configured,
		timeout: timeout,
	}, nil
}

// Check evaluates every dependency and returns the first failure with its
// stable check name. The underlying error is retained for internal startup
// diagnostics; HTTP adapters must map the result to a non-secret response.
// Readiness transitions are logged by the application lifecycle, not here.
func (a *Aggregator) Check(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	for _, check := range a.checks {
		if err := check.Run(checkCtx); err != nil {
			return fmt.Errorf("health check %q: %w", check.Name, err)
		}
	}

	return nil
}

// Ready reports whether every required dependency passes its check.
func (a *Aggregator) Ready(ctx context.Context) bool {
	return a.Check(ctx) == nil
}
