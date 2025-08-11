package controller

import (
	"context"
	"fmt"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

func (c *Controller) HealthLiveness(ctx context.Context) error {
	return nil
}

func (c *Controller) HealthReadiness(ctx context.Context) (api.HealthReadinessRes, error) {
	if c.options.ReadyState != nil && c.options.ReadyState.Load() {
		return &api.HealthReadinessOK{}, nil
	}

	return &api.HealthReadinessServiceUnavailable{}, fmt.Errorf("service unavailable")
}
