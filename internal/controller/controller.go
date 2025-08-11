package controller

import (
	"fmt"
	"sync/atomic"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

type Controller struct {
	api.Handler
	options *Options
}

type Options struct {
	ReadyState *atomic.Bool
}

func NewController(options *Options) (*Controller, error) {
	// Validate options
	if options == nil {
		return nil, fmt.Errorf("options is required")
	}

	return &Controller{
		options: options,
	}, nil
}
