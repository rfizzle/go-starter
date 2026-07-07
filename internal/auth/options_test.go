package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestWithContextKey(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	authService, err := NewService(ctx, validator, Logger(logger), ContextKey("custom_key"))
	assert.NoError(t, err)

	// Cast to concrete type to access private fields for testing
	svc := authService.(*service)
	assert.Equal(t, "custom_key", svc.contextKey)
}

func TestOptionsChaining(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	authService, err := NewService(ctx, validator,
		Logger(logger),
		ContextKey("test_context"),
	)

	assert.NoError(t, err)

	// Cast to concrete type to verify options were applied
	svc := authService.(*service)
	assert.Equal(t, "test_context", svc.contextKey)
	assert.Equal(t, logger, svc.logger)
}
