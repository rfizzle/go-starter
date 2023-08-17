package controller

import (
	"sync/atomic"

	"github.com/rfizzle/go-starter/internal/api"
)

type Controller struct {
	Health api.HealthAPI
}

func NewController(readyState *atomic.Bool) *Controller {
	return &Controller{
		Health: newHealthController(readyState),
	}
}
