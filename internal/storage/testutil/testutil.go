package testutil

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/rfizzle/go-starter/internal/storage/postgres"
	"go.uber.org/zap/zaptest"
)

// TestProvider creates a PostgreSQL provider with a mock database for testing
func TestProvider(t *testing.T) (*postgres.Provider, pgxmock.PgxPoolIface) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	provider, err := postgres.Open(ctx,
		postgres.WithMockPool(mock),
		postgres.WithLogger(logger),
	)
	if err != nil {
		t.Fatal(err)
	}

	return provider, mock
}

// TestProviderWithOptions creates a PostgreSQL provider with additional options
func TestProviderWithOptions(t *testing.T, opts ...postgres.Option) (*postgres.Provider, pgxmock.PgxPoolIface) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	allOpts := append([]postgres.Option{
		postgres.WithMockPool(mock),
		postgres.WithLogger(logger),
	}, opts...)

	provider, err := postgres.Open(ctx, allOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return provider, mock
}

// VerifyMockExpectations verifies that all mock expectations were met
func VerifyMockExpectations(t *testing.T, mock pgxmock.PgxPoolIface) {
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}

// SetupBasicExpectations sets up common mock expectations for basic operations
func SetupBasicExpectations(mock pgxmock.PgxPoolIface) {
	mock.ExpectPing()
}

// User represents a test user entity
type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

// SampleUsers returns a slice of sample users for testing
func SampleUsers() []User {
	return []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		{ID: 3, Name: "Bob Johnson", Email: "bob@example.com"},
	}
}

// MockUserQuery sets up expectations for a user query
func MockUserQuery(mock pgxmock.PgxPoolIface, users []User) {
	rows := pgxmock.NewRows([]string{"id", "name", "email"})
	for _, user := range users {
		rows.AddRow(user.ID, user.Name, user.Email)
	}

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
}

// MockUserQueryWithArgs sets up expectations for a user query with arguments
func MockUserQueryWithArgs(mock pgxmock.PgxPoolIface, users []User, args ...interface{}) {
	rows := pgxmock.NewRows([]string{"id", "name", "email"})
	for _, user := range users {
		rows.AddRow(user.ID, user.Name, user.Email)
	}

	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(args...).WillReturnRows(rows)
}

// MockUserInsert sets up expectations for a user insert
func MockUserInsert(mock pgxmock.PgxPoolIface, user User) {
	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.Name, user.Email).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
}

// MockUserUpdate sets up expectations for a user update
func MockUserUpdate(mock pgxmock.PgxPoolIface, userID int, name string) {
	mock.ExpectExec("UPDATE users SET name = (.+) WHERE id = (.+)").
		WithArgs(name, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
}

// MockUserDelete sets up expectations for a user delete
func MockUserDelete(mock pgxmock.PgxPoolIface, userID int) {
	mock.ExpectExec("DELETE FROM users WHERE id = (.+)").
		WithArgs(userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
}

// MockTransaction sets up expectations for a transaction
func MockTransaction(mock pgxmock.PgxPoolIface) {
	mock.ExpectBegin()
	mock.ExpectCommit()
}

// MockTransactionWithRollback sets up expectations for a transaction that rolls back
func MockTransactionWithRollback(mock pgxmock.PgxPoolIface) {
	mock.ExpectBegin()
	mock.ExpectRollback()
}

// MockBatchInsert sets up expectations for batch insert operations
func MockBatchInsert(mock pgxmock.PgxPoolIface, users []User) {
	mock.ExpectBegin()
	// pgxmock doesn't fully support batch operations, so we'll skip batch testing
	// or use individual exec calls in a simplified test
	for _, user := range users {
		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.Name, user.Email).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}
	mock.ExpectCommit()
}
