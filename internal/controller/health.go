package controller

import (
	"context"
	"sync/atomic"

	"github.com/go-openapi/runtime/middleware"
	"github.com/rfizzle/go-starter/internal/api"
	"github.com/rfizzle/go-starter/internal/api/operations/health"
)

type healthController struct {
	readyState *atomic.Bool
}

func newHealthController(readyState *atomic.Bool) api.HealthAPI {
	return &healthController{
		readyState: readyState,
	}
}

func (h *healthController) HealthLiveness(ctx context.Context, params health.HealthLivenessParams) middleware.Responder {
	return health.NewHealthLivenessOK()
}

func (h *healthController) HealthReadiness(ctx context.Context, params health.HealthReadinessParams) middleware.Responder {
	if h.readyState.Load() {
		return health.NewHealthReadinessOK()
	}

	return health.NewHealthReadinessServiceUnavailable()
}
