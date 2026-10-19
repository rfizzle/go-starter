package auth

import (
	"context"
)

type (
	authSubjectContextKey struct{}
)

type authSubjectContext struct {
	SubjectMap map[string]string
}

func newAuthSubjectContext(ctx context.Context, key string, subject string) (context.Context, error) {
	if subject == "" {
		return nil, &Error{
			Err: ErrEmptySubject,
		}
	}

	existingContext, exists := ctx.Value(authSubjectContextKey{}).(authSubjectContext)

	if !exists {
		return context.WithValue(ctx, authSubjectContextKey{}, authSubjectContext{
			SubjectMap: map[string]string{
				key: subject,
			},
		}), nil
	}

	// If a context does exist, copy its map before inserting so the derived
	// context does not mutate a map still visible to the parent context.
	merged := make(map[string]string, len(existingContext.SubjectMap)+1)
	for k, v := range existingContext.SubjectMap {
		merged[k] = v
	}
	merged[key] = subject
	return context.WithValue(ctx, authSubjectContextKey{}, authSubjectContext{
		SubjectMap: merged,
	}), nil
}

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

// GetSubjectFromContext returns the subject WithSubject stored under key. The
// error wraps ErrMissingSubjectContext when ctx carries no subject and
// ErrMissingKeyFromSubjectContext when it carries none under key.
func GetSubjectFromContext(ctx context.Context, key string) (string, error) {
	return subjectFromContext(ctx, key)
}

// WithSubject returns a copy of ctx carrying the authenticated subject under the
// given key. It is the public counterpart to GetSubjectFromContext, used by
// transport layers (e.g. ogen's SecurityHandler) that authenticate a request
// before it reaches a handler.
func WithSubject(ctx context.Context, key string, subject string) (context.Context, error) {
	return newAuthSubjectContext(ctx, key, subject)
}
