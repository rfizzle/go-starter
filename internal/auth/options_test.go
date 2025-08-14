package auth

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestWithContextKey(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator, WithLogger(logger), WithContextKey("custom_key"))
	assert.NoError(t, err)
	assert.Equal(t, "custom_key", auth.contextKey)
}

func TestOptionsChaining(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	auth, err := NewAuth(validator,
		WithLogger(logger),
		WithContextKey("test_context"),
	)

	assert.NoError(t, err)
	assert.Equal(t, "test_context", auth.contextKey)
	assert.Equal(t, logger, auth.logger)
}
