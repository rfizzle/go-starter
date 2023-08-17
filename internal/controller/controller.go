package controller

import (
	"github.com/rfizzle/go-starter/internal/api"
)

type Controller struct {
	Health api.HealthAPI
}

func NewController() *Controller {
	return &Controller{
		Health: newHealthController(),
	}
}
