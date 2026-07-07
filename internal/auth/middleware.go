package auth

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

// Middleware function return a middleware function that checks if the request contains a validated authorization subject
// and inserts it into the request context. If the request does not contain a validated authorization subject, the
// middleware serves the next handler without modifying the request or response.
func (s *service) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get a validated authenticated subject from request
			subject, err := s.validator.Validate(r)
			if err != nil {
				if !errors.Is(err, ErrUnauthorized) {
					s.logger.Error("request auth validation failed", zap.Error(err))
				}
				next.ServeHTTP(w, r)
				return
			}

			// Get the current request context
			ctx := r.Context()

			// Build the new context with the authenticated subject
			newCtx, err := s.newSubjectContext(ctx, subject)
			if err != nil {
				s.logger.Error("middleware: failed to build new context with authenticated subject", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			// Serve the next handler with the new context
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

// RequireAuthentication function returns a middleware function that checks if the request contains an authenticated
// subject. If the request does not contain an authenticated subject, the middleware returns a 401 Unauthorized response.
func (s *service) RequireAuthentication() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the authenticated subject from the context
			subject, err := s.subjectFromContext(r.Context())
			if err != nil {
				s.logger.Debug("authentication required but no subject in context", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Fail if the entity ID is empty
			if subject == "" {
				s.logger.Error("middleware: subject from context is empty")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthenticationWithCustomResponse allows custom error response
func (s *service) RequireAuthenticationWithCustomResponse(errorHandler func(w http.ResponseWriter, r *http.Request, err error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, err := s.subjectFromContext(r.Context())
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if subject == "" {
				errorHandler(w, r, ErrUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
