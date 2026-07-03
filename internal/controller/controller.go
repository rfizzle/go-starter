package controller

import (
	"fmt"
	"sync/atomic"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/storage"
)

// subjectContextKey is the key under which the authenticated subject is stored
// in the request context by HandleBearerAuth and read back by GetCurrentUser.
const subjectContextKey = "subject"

type Controller struct {
	api.Handler
	options *Options
}

type Options struct {
	ReadyState *atomic.Bool
	Store      storage.Client
	// Validator authenticates bearer tokens for protected operations. It may be
	// nil when the API exposes no protected operations; requests to protected
	// operations then fail closed with an unauthorized error.
	Validator auth.TokenValidator
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
