package auth

import (
	"time"

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

func WithSecurityConfig(config SecurityConfig) Options {
	return func(a *Auth) {
		a.securityConfig = &config
	}
}

func WithTokenTimeout(timeout time.Duration) Options {
	return func(a *Auth) {
		if a.securityConfig == nil {
			defaultConfig := DefaultSecurityConfig()
			a.securityConfig = &defaultConfig
		}
		a.securityConfig.TokenTimeout = timeout
	}
}

func WithMaxLoginAttempts(attempts int) Options {
	return func(a *Auth) {
		if a.securityConfig == nil {
			defaultConfig := DefaultSecurityConfig()
			a.securityConfig = &defaultConfig
		}
		a.securityConfig.MaxLoginAttempts = attempts
	}
}
