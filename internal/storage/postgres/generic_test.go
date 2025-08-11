package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestProvider_Query(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		sql       string
		args      []interface{}
		setupMock func(mock pgxmock.PgxPoolIface)
		wantErr   bool
	}{
		{
			name: "successful query",
			sql:  "SELECT id, name FROM users WHERE id = $1",
			args: []interface{}{1},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(1, "John Doe")
				mock.ExpectQuery("SELECT id, name FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query error",
			sql:  "SELECT * FROM nonexistent",
			args: []interface{}{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT \\* FROM nonexistent").
					WillReturnError(errors.New("table does not exist"))
			},
			wantErr: true,
		},
		{
			name: "query with multiple args",
			sql:  "SELECT * FROM users WHERE age > $1 AND city = $2",
			args: []interface{}{25, "New York"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "age", "city"}).
					AddRow(1, "John", 30, "New York").
					AddRow(2, "Jane", 28, "New York")
				mock.ExpectQuery("SELECT \\* FROM users WHERE age > \\$1 AND city = \\$2").
					WithArgs(25, "New York").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			tt.setupMock(mock)

			provider := &Provider{
				ctx:       ctx,
				logger:    logger,
				db:        mock,
				closeFunc: func(ctx context.Context) error { return nil },
			}

			scanner, err := provider.Query(tt.sql, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, scanner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scanner)
				assert.IsType(t, &internalScanner{}, scanner)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestProvider_Execute(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		sql       string
		args      []interface{}
		setupMock func(mock pgxmock.PgxPoolIface)
		wantErr   bool
	}{
		{
			name: "successful execute",
			sql:  "INSERT INTO users (name, email) VALUES ($1, $2)",
			args: []interface{}{"John Doe", "john@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users \\(name, email\\) VALUES \\(\\$1, \\$2\\)").
					WithArgs("John Doe", "john@example.com").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "execute error",
			sql:  "UPDATE users SET name = $1 WHERE id = $2",
			args: []interface{}{"Jane Doe", 999},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users SET name = \\$1 WHERE id = \\$2").
					WithArgs("Jane Doe", 999).
					WillReturnError(errors.New("constraint violation"))
			},
			wantErr: true,
		},
		{
			name: "delete operation",
			sql:  "DELETE FROM users WHERE id = $1",
			args: []interface{}{1},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			tt.setupMock(mock)

			provider := &Provider{
				ctx:       ctx,
				logger:    logger,
				db:        mock,
				closeFunc: func(ctx context.Context) error { return nil },
			}

			err = provider.Execute(tt.sql, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestProvider_privateQuery(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:       ctx,
		logger:    logger,
		db:        mock,
		closeFunc: func(ctx context.Context) error { return nil },
	}

	// Test successful _query
	rows := pgxmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := provider._query("SELECT COUNT(*) FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test _query error
	mock.ExpectQuery("SELECT \\* FROM invalid").
		WillReturnError(errors.New("syntax error"))

	result, err = provider._query("SELECT * FROM invalid")
	assert.Error(t, err)
	assert.Nil(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestProvider_privateExec(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:       ctx,
		logger:    logger,
		db:        mock,
		closeFunc: func(ctx context.Context) error { return nil },
	}

	// Test successful _exec
	mock.ExpectExec("UPDATE users SET active = \\$1").
		WithArgs(true).
		WillReturnResult(pgxmock.NewResult("UPDATE", 3))

	err = provider._exec("UPDATE users SET active = $1", true)
	assert.NoError(t, err)

	// Test _exec error
	mock.ExpectExec("INSERT INTO users").
		WillReturnError(errors.New("constraint violation"))

	err = provider._exec("INSERT INTO users")
	assert.Error(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestProvider_QueryExecuteIntegration(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:       ctx,
		logger:    logger,
		db:        mock,
		closeFunc: func(ctx context.Context) error { return nil },
	}

	// Simulate a complete workflow: insert, then query
	mock.ExpectExec("INSERT INTO users \\(name\\) VALUES \\(\\$1\\)").
		WithArgs("Alice").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	rows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice")
	mock.ExpectQuery("SELECT id, name FROM users WHERE name = \\$1").
		WithArgs("Alice").
		WillReturnRows(rows)

	// Execute insert
	err = provider.Execute("INSERT INTO users (name) VALUES ($1)", "Alice")
	assert.NoError(t, err)

	// Query the inserted data
	scanner, err := provider.Query("SELECT id, name FROM users WHERE name = $1", "Alice")
	assert.NoError(t, err)
	assert.NotNil(t, scanner)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
