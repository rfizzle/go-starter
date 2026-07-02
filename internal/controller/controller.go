package controller

import (
	"fmt"
	"sync/atomic"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/storage"
)

type Controller struct {
	api.Handler
	options *Options
}

type Options struct {
	ReadyState *atomic.Bool
	Store      storage.Client
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
