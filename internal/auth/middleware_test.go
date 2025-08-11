package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	middleware := auth.Middleware()

	t.Run("Valid subject", func(t *testing.T) {
		validator.Subject = "valid_subject"
		validator.Err = nil

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			subject, err := auth.subjectFromContext(ctx)
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
			subject, err := auth.subjectFromContext(ctx)
			assert.Empty(t, subject)
			assert.Error(t, err, ErrMissingSubjectContext)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)
	})
}

func TestRequireAuthentication(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	requireAuth := auth.RequireAuthentication()

	t.Run("Authenticated subject", func(t *testing.T) {
		ctx, err := auth.newSubjectContext(context.Background(), "valid_subject")
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
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})

		handler := requireAuth(nextHandler)
		handler.ServeHTTP(rr, req)
	})
}
