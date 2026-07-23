package seed

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	_, err := New(nil)
	require.Error(t, err)
}

func TestSeeder_Run_CreatesMissingUsers(t *testing.T) {
	store := newFakeUserStore()
	seeds := []UserSeed{
		{Email: "a@example.com", Password: "pw-a", FirstName: "A", LastName: "One"},
		{Email: "b@example.com", Password: "pw-b", FirstName: "B", LastName: "Two"},
	}

	s, err := New(store, WithUsers(seeds), WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	require.NoError(t, s.Run(context.Background()))

	require.Len(t, store.created, 2)
	assert.Equal(t, "a@example.com", store.created[0].Email)
	assert.Equal(t, "A", store.created[0].FirstName)

	// Password is hashed, never stored in the clear.
	assert.NotEqual(t, "pw-a", store.created[0].PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(store.created[0].PasswordHash), []byte("pw-a")))
}

func TestSeeder_Run_Idempotent(t *testing.T) {
	store := newFakeUserStore()
	seeds := []UserSeed{{Email: "a@example.com", Password: "pw-a"}}

	s, err := New(store, WithUsers(seeds), WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	require.NoError(t, s.Run(context.Background()))
	require.NoError(t, s.Run(context.Background()))
	require.NoError(t, s.Run(context.Background()))

	// Three runs, but the user is inserted exactly once.
	require.Len(t, store.created, 1)
}

func TestSeeder_Run_SkipsExisting(t *testing.T) {
	store := newFakeUserStore()
	store.byEmail["a@example.com"] = domain.User{Email: "a@example.com"}

	seeds := []UserSeed{
		{Email: "a@example.com", Password: "pw-a"},
		{Email: "b@example.com", Password: "pw-b"},
	}

	s, err := New(store, WithUsers(seeds), WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	require.NoError(t, s.Run(context.Background()))

	// Only the missing user is created.
	require.Len(t, store.created, 1)
	assert.Equal(t, "b@example.com", store.created[0].Email)
}

func TestSeeder_Run_LookupError(t *testing.T) {
	store := newFakeUserStore()
	store.byEmailErr = errors.New("boom")

	s, err := New(store, WithUsers([]UserSeed{{Email: "a@example.com"}}), WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	err = s.Run(context.Background())
	require.Error(t, err)
	assert.Empty(t, store.created)
}

func TestSeeder_Run_CreateError(t *testing.T) {
	store := newFakeUserStore()
	store.createErr = errors.New("boom")

	s, err := New(store, WithUsers([]UserSeed{{Email: "a@example.com"}}), WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	err = s.Run(context.Background())
	require.Error(t, err)
}

func TestSeeder_Run_DefaultUsers(t *testing.T) {
	store := newFakeUserStore()

	s, err := New(store, WithBcryptCost(bcrypt.MinCost))
	require.NoError(t, err)

	require.NoError(t, s.Run(context.Background()))

	require.Len(t, store.created, len(DefaultUsers))
	assert.Equal(t, DefaultUsers[0].Email, store.created[0].Email)
}
