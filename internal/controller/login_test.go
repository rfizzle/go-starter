package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

type stubAuthenticator struct {
	subject  string
	err      error
	calls    int
	email    string
	password string
}

func (s *stubAuthenticator) Authenticate(_ context.Context, email, password string) (string, error) {
	s.calls++
	s.email = email
	s.password = password
	return s.subject, s.err
}

type stubTokenGenerator struct {
	token   string
	err     error
	calls   int
	subject string
}

func (s *stubTokenGenerator) NewToken(subject string) (string, error) {
	s.calls++
	s.subject = subject
	return s.token, s.err
}

func TestLogin_MissingDependencyFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		opts func(*stubAuthenticator, *stubTokenGenerator) Options
	}{
		{
			name: "missing authenticator",
			opts: func(_ *stubAuthenticator, tokens *stubTokenGenerator) Options {
				return Options{Tokens: tokens}
			},
		},
		{
			name: "missing token generator",
			opts: func(authenticator *stubAuthenticator, _ *stubTokenGenerator) Options {
				return Options{Authenticator: authenticator}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := &stubAuthenticator{}
			tokens := &stubTokenGenerator{}
			controller := newTestController(tt.opts(authenticator, tokens))

			res, err := controller.Login(t.Context(), &api.LoginRequest{Email: "ada@example.com", Password: "password"})

			require.NoError(t, err)
			assert.IsType(t, &api.LoginUnauthorized{}, res)
			assert.Zero(t, authenticator.calls)
			assert.Zero(t, tokens.calls)
		})
	}
}

func TestLogin_InvalidCredentialsReturnsUnauthorized(t *testing.T) {
	authenticator := &stubAuthenticator{err: ErrInvalidCredentials}
	tokens := &stubTokenGenerator{}
	controller := newTestController(Options{Authenticator: authenticator, Tokens: tokens})

	res, err := controller.Login(t.Context(), &api.LoginRequest{Email: "ada@example.com", Password: "wrong-password"})

	require.NoError(t, err)
	assert.IsType(t, &api.LoginUnauthorized{}, res)
	assert.Equal(t, 1, authenticator.calls)
	assert.Zero(t, tokens.calls)
}

func TestLogin_UnexpectedAuthenticatorErrorPropagates(t *testing.T) {
	wantErr := errors.New("credential store unavailable")
	authenticator := &stubAuthenticator{err: wantErr}
	tokens := &stubTokenGenerator{}
	controller := newTestController(Options{Authenticator: authenticator, Tokens: tokens})

	res, err := controller.Login(t.Context(), &api.LoginRequest{Email: "ada@example.com", Password: "password"})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, authenticator.calls)
	assert.Zero(t, tokens.calls)
}

func TestLogin_SuccessPassesCredentialsAndSubject(t *testing.T) {
	authenticator := &stubAuthenticator{subject: "ada@example.com"}
	tokens := &stubTokenGenerator{token: "signed-token"}
	controller := newTestController(Options{Authenticator: authenticator, Tokens: tokens})

	res, err := controller.Login(t.Context(), &api.LoginRequest{Email: "ada@example.com", Password: "correct-password"})

	require.NoError(t, err)
	response, ok := res.(*api.LoginResponse)
	require.True(t, ok, "expected LoginResponse, got %T", res)
	assert.Equal(t, "signed-token", response.Token)
	assert.Equal(t, "ada@example.com", authenticator.email)
	assert.Equal(t, "correct-password", authenticator.password)
	assert.Equal(t, "ada@example.com", tokens.subject)
	assert.Equal(t, 1, authenticator.calls)
	assert.Equal(t, 1, tokens.calls)
}

func TestLogin_TokenGeneratorErrorPropagates(t *testing.T) {
	wantErr := errors.New("token signing failed")
	authenticator := &stubAuthenticator{subject: "ada@example.com"}
	tokens := &stubTokenGenerator{err: wantErr}
	controller := newTestController(Options{Authenticator: authenticator, Tokens: tokens})

	res, err := controller.Login(t.Context(), &api.LoginRequest{Email: "ada@example.com", Password: "correct-password"})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, authenticator.calls)
	assert.Equal(t, 1, tokens.calls)
}
