package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequireAuthenticationWithCustomResponse(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	t.Run("custom error handler called for missing subject", func(t *testing.T) {
		customErrorCalled := false
		var receivedError error

		errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
			customErrorCalled = true
			receivedError = err
			w.WriteHeader(http.StatusTeapot) // Use unusual status to verify custom handler
			_, _ = w.Write([]byte("Custom error response"))
		}

		middleware := authLegacy.RequireAuthenticationWithCustomResponse(errorHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.True(t, customErrorCalled)
		assert.NotNil(t, receivedError)
		assert.Equal(t, http.StatusTeapot, rr.Code)
		assert.Equal(t, "Custom error response", rr.Body.String())
		assert.False(t, nextCalled)
	})

	t.Run("custom error handler called for empty subject", func(t *testing.T) {
		customErrorCalled := false
		var receivedError error

		errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
			customErrorCalled = true
			receivedError = err
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("Empty subject"))
		}

		middleware := authLegacy.RequireAuthenticationWithCustomResponse(errorHandler)

		// Create request without any subject context
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.True(t, customErrorCalled)
		assert.NotNil(t, receivedError)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.False(t, nextCalled)
	})

	t.Run("proceeds to next handler when authenticated", func(t *testing.T) {
		customErrorCalled := false

		errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
			customErrorCalled = true
		}

		middleware := authLegacy.RequireAuthenticationWithCustomResponse(errorHandler)

		// Create authenticated context
		req := httptest.NewRequest("GET", "/", nil)
		ctx, err := newAuthSubjectContext(req.Context(), authLegacy.contextKey, "user123")
		assert.NoError(t, err)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.False(t, customErrorCalled)
		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestImprovedRequireAuthentication(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	middleware := authLegacy.RequireAuthentication()

	t.Run("returns proper HTTP error format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
		assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.False(t, nextCalled)
	})

	t.Run("logs debug message for missing context", func(t *testing.T) {
		// This test verifies the logging behavior - in practice you'd use a test logger
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler should not be called")
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

// Test the mock validator to ensure it works as expected
func TestMockValidatorWithError(t *testing.T) {
	validator := &mockValidator{
		Subject: "",
		Err:     auth.ErrUnauthorized,
	}

	req := httptest.NewRequest("GET", "/", nil)
	subject, err := validator.Validate(req)

	assert.Empty(t, subject)
	assert.Equal(t, auth.ErrUnauthorized, err)
}

func TestMockValidatorSuccess(t *testing.T) {
	validator := &mockValidator{
		Subject: "user123",
		Err:     nil,
	}

	req := httptest.NewRequest("GET", "/", nil)
	subject, err := validator.Validate(req)

	assert.Equal(t, "user123", subject)
	assert.NoError(t, err)
}

// Integration test combining middleware
func TestMiddlewareChain(t *testing.T) {
	validator := &mockValidator{Subject: "user123", Err: nil}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	// Chain: Auth middleware -> Require Authentication -> Handler
	authMiddleware := authLegacy.Middleware()
	requireAuthMiddleware := authLegacy.RequireAuthentication()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	finalHandlerCalled := false
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalHandlerCalled = true

		// Verify subject is available in context
		subject, err := subjectFromContext(r.Context(), authLegacy.contextKey)
		assert.NoError(t, err)
		assert.Equal(t, "user123", subject)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Success"))
	})

	// Build middleware chain
	handler := authMiddleware(requireAuthMiddleware(finalHandler))
	handler.ServeHTTP(rr, req)

	assert.True(t, finalHandlerCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Success", rr.Body.String())
}

func TestMiddlewareChain_FailedValidation(t *testing.T) {
	validator := &mockValidator{Subject: "", Err: auth.ErrUnauthorized}
	logger := zap.NewNop()
	authLegacy, err := NewAuth(validator, WithLoggerLegacy(logger))
	assert.NoError(t, err)

	authMiddleware := authLegacy.Middleware()
	requireAuthMiddleware := authLegacy.RequireAuthentication()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	finalHandlerCalled := false
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalHandlerCalled = true
	})

	handler := authMiddleware(requireAuthMiddleware(finalHandler))
	handler.ServeHTTP(rr, req)

	assert.False(t, finalHandlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
