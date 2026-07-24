package credential

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/domain"
)

// fakeCredentialReader is an in-memory credentialReader whose behavior a test
// controls directly. onCall, if set, runs during CredentialByEmail before it
// returns — used to simulate a reader that keeps working after its context
// is cancelled.
type fakeCredentialReader struct {
	cred   domain.Credential
	err    error
	calls  int
	onCall func()
}

func (f *fakeCredentialReader) CredentialByEmail(_ context.Context, _ string) (domain.Credential, error) {
	f.calls++
	if f.onCall != nil {
		f.onCall()
	}
	return f.cred, f.err
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func TestNewService_RequiresReader(t *testing.T) {
	_, err := NewService(nil)
	assert.Error(t, err)
}

func TestAuthenticate_Success(t *testing.T) {
	hash := hashPassword(t, "correct-password")
	reader := &fakeCredentialReader{cred: domain.Credential{UserID: "u1", Email: "ada@example.com", PasswordHash: hash}}
	svc, err := NewService(reader)
	require.NoError(t, err)

	got, err := svc.Authenticate(t.Context(), "ada@example.com", "correct-password")
	require.NoError(t, err)
	assert.Equal(t, Authenticated{UserID: "u1", Email: "ada@example.com"}, got)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	hash := hashPassword(t, "correct-password")
	reader := &fakeCredentialReader{cred: domain.Credential{UserID: "u1", Email: "ada@example.com", PasswordHash: hash}}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(t.Context(), "ada@example.com", "wrong-password")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthenticate_UnknownEmailIsIndistinguishableFromWrongPassword(t *testing.T) {
	reader := &fakeCredentialReader{err: domain.ErrNotFound}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(t.Context(), "missing@example.com", "anything")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 1, reader.calls)
}

func TestAuthenticate_MalformedHashIsInvalidCredentials(t *testing.T) {
	invalidHash := "not-a-bcrypt-hash"
	reader := &fakeCredentialReader{cred: domain.Credential{UserID: "u1", Email: "ada@example.com", PasswordHash: invalidHash}}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(t.Context(), "ada@example.com", "anything")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthenticate_RepositoryErrorPropagatesDistinctly(t *testing.T) {
	repoErr := errors.New("connection refused")
	reader := &fakeCredentialReader{err: repoErr}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(t.Context(), "ada@example.com", "anything")
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.False(t, errors.Is(err, ErrInvalidCredentials))
}

func TestAuthenticate_ReaderCancellationErrorPropagatesDistinctly(t *testing.T) {
	reader := &fakeCredentialReader{err: context.Canceled}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(t.Context(), "ada@example.com", "anything")
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, errors.Is(err, ErrInvalidCredentials))
}

func TestAuthenticate_PreCancelledContextShortCircuits(t *testing.T) {
	reader := &fakeCredentialReader{cred: domain.Credential{UserID: "u1", Email: "ada@example.com", PasswordHash: hashPassword(t, "x")}}
	svc, err := NewService(reader)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = svc.Authenticate(ctx, "ada@example.com", "x")
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, errors.Is(err, ErrInvalidCredentials))
	assert.Equal(t, 0, reader.calls, "a pre-cancelled context must short-circuit before reading")
}

func TestAuthenticate_ContextCancelledDuringReadStillDeniesSuccess(t *testing.T) {
	hash := hashPassword(t, "correct-password")
	ctx, cancel := context.WithCancel(t.Context())
	reader := &fakeCredentialReader{
		cred:   domain.Credential{UserID: "u1", Email: "ada@example.com", PasswordHash: hash},
		onCall: cancel, // simulates a reader that ignores cancellation and still returns a valid result
	}
	svc, err := NewService(reader)
	require.NoError(t, err)

	_, err = svc.Authenticate(ctx, "ada@example.com", "correct-password")
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, errors.Is(err, ErrInvalidCredentials))
}
