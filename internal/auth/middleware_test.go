package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddleware(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zap.NewNop()
	authService, err := NewService(ctx, validator, Logger(logger))
	assert.NoError(t, err)

	// Cast to concrete type to access private fields for testing
	svc := authService.(*service)
	middleware := authService.Middleware()

	t.Run("Valid subject", func(t *testing.T) {
		validator.Subject = "valid_subject"
		validator.Err = nil

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			subject, err := subjectFromContext(ctx, svc.contextKey)
			assert.Equal(t, "valid_subject", subject)
			assert.NoError(t, err)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)
	})

	t.Run("Invalid subject", func(t *testing.T) {
		validator.Subject = ""
		validator.Err = ErrUnauthorized

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			subject, err := subjectFromContext(ctx, svc.contextKey)
			assert.Empty(t, subject)
			assert.Error(t, err)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)
	})
}

func TestRequireAuthentication(t *testing.T) {
	ctx := context.Background()
	validator := &mockValidator{}
	logger := zap.NewNop()
	authService, err := NewService(ctx, validator, Logger(logger))
	assert.NoError(t, err)

	// Cast to concrete type for testing
	svc := authService.(*service)
	requireAuth := authService.RequireAuthentication()

	t.Run("Authenticated subject", func(t *testing.T) {
		ctx, err := newAuthSubjectContext(context.Background(), svc.contextKey, "valid_subject")
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.StatusOK, rr.Code)
		})

		handler := requireAuth(nextHandler)
		handler.ServeHTTP(rr, req)
	})

	t.Run("Unauthenticated subject", func(t *testing.T) {
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler should not be called")
		})

		handler := requireAuth(nextHandler)
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}