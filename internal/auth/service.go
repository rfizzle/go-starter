package auth

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// service implements the Service interface
type service struct {
	validator  Validator
	contextKey string
	logger     *zap.Logger
}

// NewService creates a new auth service implementation
func NewService(ctx context.Context, validator Validator, options ...Option) (Service, error) {
	if validator == nil {
		return nil, fmt.Errorf("validator is required")
	}

	s := &service{
		validator:  validator,
		contextKey: DefaultContextKey,
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
		return ErrEmptyContextKey
	}

	if s.logger == nil {
		return ErrEmptyLogger
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
func ContextKey(contextKey string) Option {
	return func(svc interface{}) error {
		if contextKey == "" {
			return ErrEmptyContextKey
		}
		if s, ok := svc.(*service); ok {
			s.contextKey = contextKey
		}
		return nil
	}
}

func Logger(logger *zap.Logger) Option {
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

