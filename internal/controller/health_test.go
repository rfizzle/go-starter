package controller

import (
	"context"
	"testing"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

type readinessFunc func(context.Context) bool

func (f readinessFunc) Ready(ctx context.Context) bool {
	return f(ctx)
}

func TestController_HealthReadiness(t *testing.T) {
	tests := []struct {
		name      string
		readiness ReadinessChecker
		wantOK    bool
	}{
		{
			name: "ready",
			readiness: readinessFunc(func(context.Context) bool {
				return true
			}),
			wantOK: true,
		},
		{
			name: "dependency unavailable",
			readiness: readinessFunc(func(context.Context) bool {
				return false
			}),
		},
		{name: "readiness not configured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewController(Options{Readiness: tt.readiness})

			response, err := controller.HealthReadiness(t.Context())
			if err != nil {
				t.Fatalf("HealthReadiness() error = %v, want nil stable response", err)
			}
			if tt.wantOK {
				if _, ok := response.(*api.HealthReadinessOK); !ok {
					t.Fatalf("HealthReadiness() response = %T, want *api.HealthReadinessOK", response)
				}
				return
			}
			if _, ok := response.(*api.HealthReadinessServiceUnavailable); !ok {
				t.Fatalf(
					"HealthReadiness() response = %T, want *api.HealthReadinessServiceUnavailable",
					response,
				)
			}
		})
	}
}

func TestController_HealthLivenessIgnoresDependencies(t *testing.T) {
	controller := NewController(Options{
		Readiness: readinessFunc(func(context.Context) bool {
			t.Fatal("readiness dependency evaluated by liveness")
			return false
		}),
	})

	if err := controller.HealthLiveness(t.Context()); err != nil {
		t.Fatalf("HealthLiveness() error = %v", err)
	}
}
