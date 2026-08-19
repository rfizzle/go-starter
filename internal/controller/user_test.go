package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/domain"
)

type userReaderFunc func(context.Context, string) (domain.User, error)

func (f userReaderFunc) ByEmail(ctx context.Context, email string) (domain.User, error) {
	return f(ctx, email)
}

func TestGetCurrentUser_EnrichesSubjectFromUserReader(t *testing.T) {
	const subject = "ada@example.com"

	reader := userReaderFunc(func(_ context.Context, email string) (domain.User, error) {
		assert.Equal(t, subject, email)
		return domain.User{
			Email:     email,
			FirstName: "Ada",
			LastName:  "Lovelace",
		}, nil
	})
	c := newTestController(Options{
		Users:     reader,
		Validator: stubValidator{subject: subject},
	})
	ctx, err := c.HandleBearerAuth(
		context.Background(),
		api.GetCurrentUserOperation,
		api.BearerAuth{Token: "valid"},
	)
	require.NoError(t, err)

	res, err := c.GetCurrentUser(ctx)
	require.NoError(t, err)

	user, ok := res.(*api.CurrentUserResponse)
	require.True(t, ok, "expected CurrentUserResponse, got %T", res)
	assert.Equal(t, subject, user.Subject)
	assert.Equal(t, subject, user.Email.Or(""))
	assert.Equal(t, "Ada", user.FirstName.Or(""))
	assert.Equal(t, "Lovelace", user.LastName.Or(""))
}

func TestGetCurrentUser_MissingStoredUserReturnsSubject(t *testing.T) {
	const subject = "missing@example.com"

	c := newTestController(Options{
		Users: userReaderFunc(func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{}, domain.ErrNotFound
		}),
		Validator: stubValidator{subject: subject},
	})
	ctx, err := c.HandleBearerAuth(
		context.Background(),
		api.GetCurrentUserOperation,
		api.BearerAuth{Token: "valid"},
	)
	require.NoError(t, err)

	res, err := c.GetCurrentUser(ctx)
	require.NoError(t, err)

	user, ok := res.(*api.CurrentUserResponse)
	require.True(t, ok, "expected CurrentUserResponse, got %T", res)
	assert.Equal(t, subject, user.Subject)
	assert.False(t, user.Email.IsSet())
	assert.False(t, user.FirstName.IsSet())
	assert.False(t, user.LastName.IsSet())
}

func TestGetCurrentUser_UserReaderFailurePropagates(t *testing.T) {
	wantErr := errors.New("read user")

	c := newTestController(Options{
		Users: userReaderFunc(func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{}, wantErr
		}),
		Validator: stubValidator{subject: "ada@example.com"},
	})
	ctx, err := c.HandleBearerAuth(
		context.Background(),
		api.GetCurrentUserOperation,
		api.BearerAuth{Token: "valid"},
	)
	require.NoError(t, err)

	res, err := c.GetCurrentUser(ctx)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, wantErr)
}
