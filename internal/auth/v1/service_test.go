package v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/rfizzle/go-starter/internal/auth"
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
	assert.Equal(t, auth.ErrEmptyContextKey, err)

	// Test case: Missing validator
	service, err = NewService(ctx, nil, Logger(logger))
	assert.Nil(t, service)
	assert.Error(t, err)
}

func TestNewAuth(t *testing.T) {
	// Test case: Successful creation of Auth
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NotNil(t, authLegacy)
	assert.Nil(t, err)
	assert.Equal(t, auth.DefaultContextKey, authLegacy.contextKey)
	assert.Equal(t, logger, authLegacy.Logger())

	// Test case: Missing context key
	authLegacy, err = NewAuth(validator, WithContextKeyLegacy(""))
	assert.Nil(t, authLegacy)
	assert.Equal(t, auth.ErrEmptyContextKey, err)

	// Test case: Missing logger
	authLegacy, err = NewAuth(validator)
	assert.Nil(t, authLegacy)
	assert.Equal(t, auth.ErrEmptyLogger, err)
}

func TestServiceInterface(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)
	
	service, err := NewService(ctx, validator, Logger(logger))
	assert.Nil(t, err)
	assert.NotNil(t, service)
	
	// Test that service implements auth.Service interface
	var _ auth.Service = service
	
	// Test that service has all required methods
	assert.NotNil(t, service.Middleware())
	assert.NotNil(t, service.RequireAuthentication())
	assert.NotNil(t, service.RequireAuthenticationWithCustomResponse(nil))
	assert.NotNil(t, service.Logger())
}