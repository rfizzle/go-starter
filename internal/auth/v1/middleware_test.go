package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddleware(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	middleware := authLegacy.Middleware()

	t.Run("Valid subject", func(t *testing.T) {
		validator.Subject = "valid_subject"
		validator.Err = nil

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			subject, err := subjectFromContext(ctx, authLegacy.contextKey)
			assert.Equal(t, "valid_subject", subject)
			assert.NoError(t, err)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)
	})

	t.Run("Invalid subject", func(t *testing.T) {
		validator.Subject = ""
		validator.Err = auth.ErrUnauthorized

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			subject, err := subjectFromContext(ctx, authLegacy.contextKey)
			assert.Empty(t, subject)
			assert.Error(t, err)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)
	})
}

func TestRequireAuthentication(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	requireAuth := authLegacy.RequireAuthentication()

	t.Run("Authenticated subject", func(t *testing.T) {
		ctx, err := newAuthSubjectContext(context.Background(), authLegacy.contextKey, "valid_subject")
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