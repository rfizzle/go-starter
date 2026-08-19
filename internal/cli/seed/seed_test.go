package seed

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/domain"
)

// fakeUserStore is an in-memory UserStore keyed by email. It records Create
// calls so tests can assert both the resulting state and what was inserted.
type fakeUserStore struct {
	byEmail    map[string]domain.User
	created    []domain.CreateUserParams
	byEmailErr error
	createErr  error
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{byEmail: map[string]domain.User{}}
}

func (f *fakeUserStore) ByEmail(_ context.Context, email string) (domain.User, error) {
	if f.byEmailErr != nil {
		return domain.User{}, f.byEmailErr
	}
	u, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserStore) Create(_ context.Context, params domain.CreateUserParams) (domain.User, error) {
	if f.createErr != nil {
		return domain.User{}, f.createErr
	}
	f.created = append(f.created, params)
	u := domain.User{Email: params.Email, FirstName: params.FirstName, LastName: params.LastName}
	f.byEmail[params.Email] = u
	return u, nil
}

func TestNew_NilStore(t *testing.T) {
	_, err := New(nil, zap.NewNop(), []UserSeed{{Email: "a@example.com", Password: "pw-a"}})
	require.Error(t, err)
}

func TestNew_RequiresExplicitUsersAndCredentials(t *testing.T) {
	store := newFakeUserStore()

	tests := []struct {
		name  string
		seeds []UserSeed
	}{
		{name: "no users"},
		{name: "missing email", seeds: []UserSeed{{Password: "password"}}},
		{name: "missing password", seeds: []UserSeed{{Email: "a@example.com"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(store, zap.NewNop(), tt.seeds)
			require.Error(t, err)
		})
	}
}

func TestSeeder_Run_CreatesMissingUsers(t *testing.T) {
	store := newFakeUserStore()
	seeds := []UserSeed{
		{Email: "a@example.com", Password: "pw-a", FirstName: "A", LastName: "One"},
		{Email: "b@example.com", Password: "pw-b", FirstName: "B", LastName: "Two"},
	}

	s, err := New(store, zap.NewNop(), seeds, WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	result, err := s.Run(context.Background())
	require.NoError(t, err)

	require.Len(t, store.created, 2)
	assert.Equal(t, Result{Created: 2}, result)
	assert.Equal(t, "a@example.com", store.created[0].Email)
	assert.Equal(t, "A", store.created[0].FirstName)

	// Password is hashed, never stored in the clear.
	assert.NotEqual(t, "pw-a", store.created[0].PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(store.created[0].PasswordHash), []byte("pw-a")))
}

func TestSeeder_Run_Idempotent(t *testing.T) {
	store := newFakeUserStore()
	seeds := []UserSeed{{Email: "a@example.com", Password: "pw-a"}}

	s, err := New(store, zap.NewNop(), seeds, WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	first, err := s.Run(context.Background())
	require.NoError(t, err)
	second, err := s.Run(context.Background())
	require.NoError(t, err)
	third, err := s.Run(context.Background())
	require.NoError(t, err)

	// Three runs, but the user is inserted exactly once.
	require.Len(t, store.created, 1)
	assert.Equal(t, Result{Created: 1}, first)
	assert.Equal(t, Result{Existing: 1}, second)
	assert.Equal(t, Result{Existing: 1}, third)
}

func TestSeeder_Run_SkipsExisting(t *testing.T) {
	store := newFakeUserStore()
	store.byEmail["a@example.com"] = domain.User{Email: "a@example.com"}

	seeds := []UserSeed{
		{Email: "a@example.com", Password: "pw-a"},
		{Email: "b@example.com", Password: "pw-b"},
	}

	s, err := New(store, zap.NewNop(), seeds, WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	result, err := s.Run(context.Background())
	require.NoError(t, err)

	// Only the missing user is created.
	require.Len(t, store.created, 1)
	assert.Equal(t, "b@example.com", store.created[0].Email)
	assert.Equal(t, Result{Created: 1, Existing: 1}, result)
}

func TestSeeder_Run_LookupError(t *testing.T) {
	store := newFakeUserStore()
	store.byEmailErr = errors.New("boom")

	s, err := New(
		store,
		zap.NewNop(),
		[]UserSeed{{Email: "a@example.com", Password: "password"}},
		WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)

	_, err = s.Run(context.Background())
	require.Error(t, err)
	assert.Empty(t, store.created)
}

func TestSeeder_Run_CreateError(t *testing.T) {
	store := newFakeUserStore()
	store.createErr = errors.New("boom")

	s, err := New(
		store,
		zap.NewNop(),
		[]UserSeed{{Email: "a@example.com", Password: "password"}},
		WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)

	_, err = s.Run(context.Background())
	require.Error(t, err)
}

func TestSeeder_Run_DoesNotLogPlaintextPassword(t *testing.T) {
	store := newFakeUserStore()
	plaintext := strings.Join([]string{"plaintext", "must", "not", "be", "logged"}, "-")
	var logs bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&logs),
		zap.DebugLevel,
	))

	s, err := New(
		store,
		logger,
		[]UserSeed{{Email: "a@example.com", Password: plaintext}},
		WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)

	_, err = s.Run(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, logs.String(), plaintext)
}
