package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/rfizzle/go-starter/internal/api"
	"github.com/rfizzle/go-starter/internal/api/operations/health"
)

type healthController struct {
}

func newHealthController() api.HealthAPI {
	return &healthController{}
}

func (h *healthController) HealthLiveness(ctx context.Context, params health.HealthLivenessParams) middleware.Responder {
	//TODO implement me
	panic("implement me")
}

func (h *healthController) HealthReadiness(ctx context.Context, params health.HealthReadinessParams) middleware.Responder {
	//TODO implement me
	panic("implement me")
}
