package auth

import (
	"go.uber.org/zap"
)

type Options func(a *Auth)

func WithContextKey(contextKey string) Options {
	return func(a *Auth) {
		a.contextKey = contextKey
	}
}

func WithLogger(logger *zap.Logger) Options {
	return func(a *Auth) {
		a.logger = logger
	}
}
