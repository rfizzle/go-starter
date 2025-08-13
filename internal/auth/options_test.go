package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestWithContextKey(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator, WithLogger(logger), WithContextKey("custom_key"))
	assert.NoError(t, err)
	assert.Equal(t, "custom_key", auth.contextKey)
}

func TestWithSecurityConfig(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	customConfig := SecurityConfig{
		TokenTimeout:     12 * time.Hour,
		MaxLoginAttempts: 3,
		LockoutDuration:  10 * time.Minute,
		RequireHTTPS:     false,
		SameSitePolicy:   "Lax",
	}

	auth, err := NewAuth(validator, WithLogger(logger), WithSecurityConfig(customConfig))
	assert.NoError(t, err)
	assert.NotNil(t, auth.securityConfig)
	assert.Equal(t, 12*time.Hour, auth.securityConfig.TokenTimeout)
	assert.Equal(t, 3, auth.securityConfig.MaxLoginAttempts)
	assert.Equal(t, 10*time.Minute, auth.securityConfig.LockoutDuration)
	assert.False(t, auth.securityConfig.RequireHTTPS)
	assert.Equal(t, "Lax", auth.securityConfig.SameSitePolicy)
}

func TestWithTokenTimeout(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator, WithLogger(logger), WithTokenTimeout(6*time.Hour))
	assert.NoError(t, err)
	assert.NotNil(t, auth.securityConfig)
	assert.Equal(t, 6*time.Hour, auth.securityConfig.TokenTimeout)
	// Should use defaults for other values
	assert.Equal(t, 5, auth.securityConfig.MaxLoginAttempts)
	assert.Equal(t, 15*time.Minute, auth.securityConfig.LockoutDuration)
}

func TestWithMaxLoginAttempts(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator, WithLogger(logger), WithMaxLoginAttempts(10))
	assert.NoError(t, err)
	assert.NotNil(t, auth.securityConfig)
	assert.Equal(t, 10, auth.securityConfig.MaxLoginAttempts)
	// Should use defaults for other values
	assert.Equal(t, 24*time.Hour, auth.securityConfig.TokenTimeout)
	assert.Equal(t, 15*time.Minute, auth.securityConfig.LockoutDuration)
}

func TestMultipleSecurityOptions(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator,
		WithLogger(logger),
		WithTokenTimeout(8*time.Hour),
		WithMaxLoginAttempts(7),
	)
	assert.NoError(t, err)
	assert.NotNil(t, auth.securityConfig)
	assert.Equal(t, 8*time.Hour, auth.securityConfig.TokenTimeout)
	assert.Equal(t, 7, auth.securityConfig.MaxLoginAttempts)
	// Should use defaults for other values
	assert.Equal(t, 15*time.Minute, auth.securityConfig.LockoutDuration)
	assert.True(t, auth.securityConfig.RequireHTTPS)
}

func TestOptionsChaining(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator,
		WithLogger(logger),
		WithContextKey("test_context"),
		WithTokenTimeout(2*time.Hour),
		WithMaxLoginAttempts(3),
	)

	assert.NoError(t, err)
	assert.Equal(t, "test_context", auth.contextKey)
	assert.Equal(t, logger, auth.logger)
	assert.NotNil(t, auth.securityConfig)
	assert.Equal(t, 2*time.Hour, auth.securityConfig.TokenTimeout)
	assert.Equal(t, 3, auth.securityConfig.MaxLoginAttempts)
}
