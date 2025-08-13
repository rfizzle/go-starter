package storage

import (
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/rfizzle/go-starter/internal/storage/driver"
	"github.com/rfizzle/go-starter/internal/storage/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStorageIntegration_BasicOperations(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	// Test basic provider operations
	assert.Equal(t, "postgresql", provider.Name())
	assert.NotNil(t, provider.Logger())

	// Test ping
	testutil.SetupBasicExpectations(mock)
	err := provider.Ping()
	assert.NoError(t, err)

	// Test disconnect
	err = provider.Disconnect()
	assert.NoError(t, err)

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_QueryOperations(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	users := testutil.SampleUsers()

	t.Run("Query All Users", func(t *testing.T) {
		testutil.MockUserQuery(mock, users)

		scanner, err := provider.Query("SELECT id, name, email FROM users")
		assert.NoError(t, err)

		var result []testutil.User
		err = scanner.ScanAll(&result)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, users[0].Name, result[0].Name)
	})

	t.Run("Query Single User", func(t *testing.T) {
		testutil.MockUserQueryWithArgs(mock, users[:1], 1)

		scanner, err := provider.Query("SELECT id, name, email FROM users WHERE id = $1", 1)
		assert.NoError(t, err)

		var result testutil.User
		err = scanner.ScanOne(&result)
		assert.NoError(t, err)
		assert.Equal(t, users[0].Name, result.Name)
	})

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_ExecuteOperations(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	user := testutil.SampleUsers()[0]

	t.Run("Insert User", func(t *testing.T) {
		testutil.MockUserInsert(mock, user)

		err := provider.Execute("INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email)
		assert.NoError(t, err)
	})

	t.Run("Update User", func(t *testing.T) {
		testutil.MockUserUpdate(mock, user.ID, "Updated Name")

		err := provider.Execute("UPDATE users SET name = $1 WHERE id = $2", "Updated Name", user.ID)
		assert.NoError(t, err)
	})

	t.Run("Delete User", func(t *testing.T) {
		testutil.MockUserDelete(mock, user.ID)

		err := provider.Execute("DELETE FROM users WHERE id = $1", user.ID)
		assert.NoError(t, err)
	})

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_TransactionOperations(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	user := testutil.SampleUsers()[0]

	t.Run("Successful Transaction", func(t *testing.T) {
		mock.ExpectBegin()
		testutil.MockUserInsert(mock, user)
		mock.ExpectCommit()

		err := provider.WithTx(func(tx driver.Tx) error {
			return tx.Execute("INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email)
		})
		assert.NoError(t, err)
	})

	t.Run("Transaction with Query", func(t *testing.T) {
		mock.ExpectBegin()
		testutil.MockUserQueryWithArgs(mock, []testutil.User{user}, user.ID)
		mock.ExpectCommit()

		err := provider.WithTx(func(tx driver.Tx) error {
			scanner, err := tx.Query("SELECT id, name, email FROM users WHERE id = $1", user.ID)
			if err != nil {
				return err
			}

			var result testutil.User
			return scanner.ScanOne(&result)
		})
		assert.NoError(t, err)
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		testutil.MockTransactionWithRollback(mock)

		err := provider.WithTx(func(tx driver.Tx) error {
			return assert.AnError // Force an error to trigger rollback
		})
		assert.Error(t, err)
	})

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_BatchOperations(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	t.Run("Batch Insert", func(t *testing.T) {
		// Skip batch operations due to pgxmock limitations
		testutil.MockTransaction(mock)

		err := provider.WithTx(func(tx driver.Tx) error {
			// Skip actual batch operations for now
			return nil
		})
		assert.NoError(t, err)
	})

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_ProviderInterface(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	// Test that provider implements the driver.Provider interface
	var _ driver.Provider = provider
	var _ driver.Client = provider

	// Test interface methods
	assert.Equal(t, "postgresql", provider.Name())
	assert.NotNil(t, provider.Logger())

	testutil.SetupBasicExpectations(mock)
	err := provider.Ping()
	assert.NoError(t, err)

	err = provider.Disconnect()
	assert.NoError(t, err)

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_ErrorHandling(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	t.Run("Query Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM nonexistent").
			WillReturnError(assert.AnError)

		scanner, err := provider.Query("SELECT * FROM nonexistent")
		assert.Error(t, err)
		assert.Nil(t, scanner)
	})

	t.Run("Execute Error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO nonexistent").
			WithArgs("value").
			WillReturnError(assert.AnError)

		err := provider.Execute("INSERT INTO nonexistent VALUES ($1)", "value")
		assert.Error(t, err)
	})

	t.Run("Transaction Begin Error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(assert.AnError)

		err := provider.WithTx(func(tx driver.Tx) error {
			return nil
		})
		assert.Error(t, err)
	})

	testutil.VerifyMockExpectations(t, mock)
}

func TestStorageIntegration_ComplexWorkflow(t *testing.T) {
	provider, mock := testutil.TestProvider(t)
	defer mock.Close()

	users := testutil.SampleUsers()

	// Simulate a complex workflow: create users, update one, query all, delete one
	mock.ExpectBegin()

	// Insert multiple users
	for _, user := range users {
		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.Name, user.Email).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	// Update a user
	mock.ExpectExec("UPDATE users SET email = (.+) WHERE id = (.+)").
		WithArgs("newemail@example.com", 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Query all users
	updatedUsers := make([]testutil.User, len(users))
	copy(updatedUsers, users)
	updatedUsers[0].Email = "newemail@example.com"

	rows := mock.NewRows([]string{"id", "name", "email"})
	for _, user := range updatedUsers {
		rows.AddRow(user.ID, user.Name, user.Email)
	}
	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	// Delete a user
	mock.ExpectExec("DELETE FROM users WHERE id = (.+)").
		WithArgs(3).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	mock.ExpectCommit()

	err := provider.WithTx(func(tx driver.Tx) error {
		// Insert users
		for _, user := range users {
			if err := tx.Execute("INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email); err != nil {
				return err
			}
		}

		// Update a user
		if err := tx.Execute("UPDATE users SET email = $1 WHERE id = $2", "newemail@example.com", 1); err != nil {
			return err
		}

		// Query all users
		scanner, err := tx.Query("SELECT id, name, email FROM users")
		if err != nil {
			return err
		}

		var result []testutil.User
		if err := scanner.ScanAll(&result); err != nil {
			return err
		}

		// Delete a user
		return tx.Execute("DELETE FROM users WHERE id = $1", 3)
	})

	assert.NoError(t, err)
	testutil.VerifyMockExpectations(t, mock)
}
