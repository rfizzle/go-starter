package auth

import (
	"net/http"

	"go.uber.org/zap"
)

const (
	DefaultContextKey = "default"
)

type Generator interface {
	NewToken(subject string) (string, error)
	NewCookie(subject string) (*http.Cookie, error)
}

type Validator interface {
	Validate(r *http.Request) (string, error)
}

type Auth struct {
	validator  Validator
	contextKey string
	logger     *zap.Logger
}

func NewAuth(validator Validator, options ...Options) (*Auth, error) {
	// Set up auth
	a := &Auth{
		validator:  validator,
		contextKey: DefaultContextKey,
		logger:     nil,
	}

	// Apply options
	for _, option := range options {
		option(a)
	}

	// Validate auth
	err := a.validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Auth) validate() error {
	// Check if context key is set
	if a.contextKey == "" {
		return ErrEmptyContextKey
	}

	if a.logger == nil {
		return ErrEmptyLogger
	}

	return nil
}

func (a *Auth) Logger() *zap.Logger {
	return a.logger
}
