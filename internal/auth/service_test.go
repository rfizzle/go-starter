package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type mockValidator struct {
	Subject string
	Err     error
}

func (mV *mockValidator) Validate(r *http.Request) (string, error) {
	return mV.Subject, mV.Err
}

func TestNewService(t *testing.T) {
	ctx := context.Background()

	// Test case: Successful creation of Service
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)
	service, err := NewService(ctx, validator, Logger(logger))
	assert.NotNil(t, service)
	assert.Nil(t, err)

	// Test case: Missing context key
	service, err = NewService(ctx, validator, ContextKey(""))
	assert.Nil(t, service)
	assert.Equal(t, ErrEmptyContextKey, err)

	// Test case: Missing validator
	service, err = NewService(ctx, nil, Logger(logger))
	assert.Nil(t, service)
	assert.Error(t, err)
}

// TestNewAuth is removed as we've consolidated to NewService only

func TestServiceInterface(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)

	service, err := NewService(ctx, validator, Logger(logger))
	assert.Nil(t, err)
	assert.NotNil(t, service)

	// Test that service has all required methods
	assert.NotNil(t, service.Middleware())
	assert.NotNil(t, service.RequireAuthentication())
	assert.NotNil(t, service.RequireAuthenticationWithCustomResponse(nil))
	assert.NotNil(t, service.Logger())
}
