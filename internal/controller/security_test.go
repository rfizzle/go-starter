package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

// stubValidator is a TokenValidator whose behaviour the test controls.
type stubValidator struct {
	subject string
	err     error
}

func (s stubValidator) ValidateToken(token string) (string, error) {
	return s.subject, s.err
}

func newTestController(opts Options) *Controller {
	return NewController(opts)
}

func TestHandleBearerAuth_NoValidatorFailsClosed(t *testing.T) {
	c := newTestController(Options{})

	_, err := c.HandleBearerAuth(context.Background(), api.GetCurrentUserOperation, api.BearerAuth{Token: "anything"})
	assert.Error(t, err)
}

func TestHandleBearerAuth_InvalidTokenPropagatesError(t *testing.T) {
	wantErr := errors.New("bad token")
	c := newTestController(Options{Validator: stubValidator{err: wantErr}})

	_, err := c.HandleBearerAuth(context.Background(), api.GetCurrentUserOperation, api.BearerAuth{Token: "bad"})
	assert.ErrorIs(t, err, wantErr)
}

func TestHandleBearerAuth_ValidTokenInjectsSubject(t *testing.T) {
	c := newTestController(Options{Validator: stubValidator{subject: "user-123"}})

	ctx, err := c.HandleBearerAuth(context.Background(), api.GetCurrentUserOperation, api.BearerAuth{Token: "good"})
	require.NoError(t, err)

	// GetCurrentUser reads the subject placed on the context above.
	res, err := c.GetCurrentUser(ctx)
	require.NoError(t, err)

	user, ok := res.(*api.CurrentUserResponse)
	require.True(t, ok, "expected CurrentUserResponse, got %T", res)
	assert.Equal(t, "user-123", user.Subject)
}

func TestGetCurrentUser_NoSubjectUnauthorized(t *testing.T) {
	c := newTestController(Options{})

	res, err := c.GetCurrentUser(context.Background())
	require.NoError(t, err)

	_, ok := res.(*api.GetCurrentUserUnauthorized)
	assert.True(t, ok, "expected Unauthorized, got %T", res)
}
