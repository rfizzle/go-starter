package v1

import (
	"context"
	"fmt"

	"github.com/rfizzle/go-starter/internal/auth"
	"go.uber.org/zap"
)

// service implements the auth.Service interface
type service struct {
	validator  auth.Validator
	contextKey string
	logger     *zap.Logger
}

// NewService creates a new auth service implementation
func NewService(ctx context.Context, validator auth.Validator, options ...auth.Option) (auth.Service, error) {
	if validator == nil {
		return nil, fmt.Errorf("validator is required")
	}

	s := &service{
		validator:  validator,
		contextKey: auth.DefaultContextKey,
		logger:     zap.NewNop(),
	}

	// Apply options
	for _, option := range options {
		err := option(s)
		if err != nil {
			return nil, err
		}
	}

	// Validate
	err := s.validate()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *service) validate() error {
	if s.contextKey == "" {
		return auth.ErrEmptyContextKey
	}

	if s.logger == nil {
		return auth.ErrEmptyLogger
	}

	if s.validator == nil {
		return fmt.Errorf("validator is required")
	}

	return nil
}

func (s *service) Logger() interface{} {
	return s.logger
}

// Option functions for configuring the service
func ContextKey(contextKey string) auth.Option {
	return func(svc interface{}) error {
		if contextKey == "" {
			return auth.ErrEmptyContextKey
		}
		if s, ok := svc.(*service); ok {
			s.contextKey = contextKey
		}
		return nil
	}
}

func Logger(logger *zap.Logger) auth.Option {
	return func(svc interface{}) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}
		if s, ok := svc.(*service); ok {
			s.logger = logger
		}
		return nil
	}
}

// Legacy compatibility functions
type LegacyOptions func(*legacyAuth)

type legacyAuth struct {
	validator  auth.Validator
	contextKey string
	logger     *zap.Logger
}

// NewAuth creates a new Auth provider (legacy compatibility function)
func NewAuth(validator auth.Validator, options ...LegacyOptions) (*legacyAuth, error) {
	if validator == nil {
		return nil, fmt.Errorf("validator is required")
	}

	a := &legacyAuth{
		validator:  validator,
		contextKey: auth.DefaultContextKey,
		logger:     nil, // Don't set default for legacy compatibility
	}

	// Apply legacy options
	for _, option := range options {
		option(a)
	}

	// Check for empty context key first (this can be set by options)
	if a.contextKey == "" {
		return nil, auth.ErrEmptyContextKey
	}

	// If no logger was provided via options, return error
	if a.logger == nil {
		return nil, auth.ErrEmptyLogger
	}

	return a, nil
}

func WithContextKeyLegacy(contextKey string) LegacyOptions {
	return func(a *legacyAuth) {
		a.contextKey = contextKey
	}
}

func WithLoggerLegacy(logger *zap.Logger) LegacyOptions {
	return func(a *legacyAuth) {
		a.logger = logger
	}
}

func (a *legacyAuth) Logger() *zap.Logger {
	return a.logger
}
