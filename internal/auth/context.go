package auth

import (
	"context"
)

type (
	authSubjectContextKey struct{}
)

// A authSubjectContext is a struct that contains a subject map.
type authSubjectContext struct {
	SubjectMap map[string]string
}

// newAuthSubjectContext creates a new context for a Subject.
func newAuthSubjectContext(ctx context.Context, key string, subject string) (context.Context, error) {
	if subject == "" {
		return nil, &Error{
			Err: ErrEmptySubject,
		}
	}

	// Get existing context if it exists
	existingContext, exists := ctx.Value(authSubjectContextKey{}).(authSubjectContext)

	// If a context does not exist, create a new one and return
	if !exists {
		return context.WithValue(ctx, authSubjectContextKey{}, authSubjectContext{
			SubjectMap: map[string]string{
				key: subject,
			},
		}), nil
	}

	// If a context does exist, add the new Subject to the map and return
	existingContext.SubjectMap[key] = subject
	return context.WithValue(ctx, authSubjectContextKey{}, existingContext), nil
}

// subjectFromContext returns a context with a subject map if present.
func subjectFromContext(ctx context.Context, key string) (string, error) {
	v, ok := ctx.Value(authSubjectContextKey{}).(authSubjectContext)
	if !ok {
		return "", &Error{
			Err: ErrMissingSubjectContext,
		}
	}

	s, ok := v.SubjectMap[key]
	if !ok {
		return "", &Error{
			Err: ErrMissingKeyFromSubjectContext,
		}
	}

	return s, nil
}

// newSubjectContext creates a new context for a subject using the service contextKey.
func (s *service) newSubjectContext(ctx context.Context, subject string) (context.Context, error) {
	return newAuthSubjectContext(ctx, s.contextKey, subject)
}

// subjectFromContext returns a context with a subject map if present using the service contextKey.
func (s *service) subjectFromContext(ctx context.Context) (string, error) {
	return subjectFromContext(ctx, s.contextKey)
}

// GetSubjectFromContext is a public function to get subject from context with a specific key
func GetSubjectFromContext(ctx context.Context, key string) (string, error) {
	return subjectFromContext(ctx, key)
}