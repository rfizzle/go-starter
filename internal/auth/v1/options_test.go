package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestWithContextKey(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger), WithContextKeyLegacy("custom_key"))
	assert.NoError(t, err)
	assert.Equal(t, "custom_key", authLegacy.contextKey)
}

func TestOptionsChaining(t *testing.T) {
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	authLegacy, err := NewAuth(validator,
		WithLoggerLegacy(logger),
		WithContextKeyLegacy("test_context"),
	)

	assert.NoError(t, err)
	assert.Equal(t, "test_context", authLegacy.contextKey)
	assert.Equal(t, logger, authLegacy.logger)
}