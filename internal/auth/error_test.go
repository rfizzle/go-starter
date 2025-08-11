package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetError(t *testing.T) {
	// Test case: Error is nil
	authErr, err := GetError(nil)
	assert.Nil(t, authErr)
	assert.NotNil(t, err)

	// Test case: Error is not nil
	expectedErr := Error{Err: ErrEmptyContextKey}
	castError := error(&expectedErr)
	authErr, err = GetError(castError)
	assert.NotNil(t, authErr)
	assert.Nil(t, err)
	assert.Equal(t, &expectedErr, authErr)
	assert.Equal(t, ErrEmptyContextKey, authErr.Err)
}

func TestAuthError_Error(t *testing.T) {
	// Test case: AuthError with a specific error
	err := Error{Err: ErrEmptyLogger}

	assert.Equal(t, err.Error(), fmt.Sprintf("auth error: %s", ErrEmptyLogger.Error()))
}
