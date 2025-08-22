package auth

import (
	"context"
)

// GetSubjectFromContext is a convenience function to get subject from context
// This is deprecated - import auth/v1 directly instead
//
// Deprecated: Use v1.GetSubjectFromContext instead
func GetSubjectFromContext(ctx context.Context, key string) (string, error) {
	// Consumers should import auth/v1 directly to avoid circular imports
	return "", ErrMissingSubjectContext
}