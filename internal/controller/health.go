package controller

import (
	"context"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

func (c *Controller) HealthLiveness(ctx context.Context) error {
	return nil
}

func (c *Controller) HealthReadiness(ctx context.Context) (api.HealthReadinessRes, error) {
	if c.options.Readiness != nil && c.options.Readiness.Ready(ctx) {
		return &api.HealthReadinessOK{}, nil
	}

	return &api.HealthReadinessServiceUnavailable{}, nil
}
