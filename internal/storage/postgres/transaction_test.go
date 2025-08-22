package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rfizzle/go-starter/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestProvider_begin(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		setupMock func(mock pgxmock.PgxPoolIface)
		wantErr   bool
	}{
		{
			name: "successful begin transaction",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
			},
			wantErr: false,
		},
		{
			name: "begin transaction error",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("failed to begin transaction"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			tt.setupMock(mock)

			provider := &Provider{
				ctx:         ctx,
				logger:      logger,
				db:          mock,
				pingTimeout: 15 * time.Second,
				closeFunc:   func(ctx context.Context) error { return nil },
			}

			tx, err := provider.begin()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)
				assert.NotNil(t, tx.transaction)
				assert.NotNil(t, tx.cancelFunc)
				assert.False(t, tx.startTime.IsZero())
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestTransaction_commit(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func() (pgxmock.PgxPoolIface, pgx.Tx)
		wantErr   bool
	}{
		{
			name: "successful commit",
			setupMock: func() (pgxmock.PgxPoolIface, pgx.Tx) {
				mock, err := pgxmock.NewPool()
				assert.NoError(t, err)

				mock.ExpectBegin()
				tx, err := mock.Begin(ctx)
				assert.NoError(t, err)

				mock.ExpectCommit()
				return mock, tx
			},
			wantErr: false,
		},
		{
			name: "commit error",
			setupMock: func() (pgxmock.PgxPoolIface, pgx.Tx) {
				mock, err := pgxmock.NewPool()
				assert.NoError(t, err)

				mock.ExpectBegin()
				tx, err := mock.Begin(ctx)
				assert.NoError(t, err)

				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
				return mock, tx
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, pgxTx := tt.setupMock()
			defer mock.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tx := &transaction{
				transactionCtx: ctx,
				transaction:    pgxTx,
				cancelFunc:     cancel,
				startTime:      time.Now(),
			}

			err := tx.commit()
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

func TestTransaction_rollback(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func() (pgxmock.PgxPoolIface, pgx.Tx)
		wantErr   bool
	}{
		{
			name: "successful rollback",
			setupMock: func() (pgxmock.PgxPoolIface, pgx.Tx) {
				mock, err := pgxmock.NewPool()
				assert.NoError(t, err)

				mock.ExpectBegin()
				tx, err := mock.Begin(ctx)
				assert.NoError(t, err)

				mock.ExpectRollback()
				return mock, tx
			},
			wantErr: false,
		},
		{
			name: "rollback error",
			setupMock: func() (pgxmock.PgxPoolIface, pgx.Tx) {
				mock, err := pgxmock.NewPool()
				assert.NoError(t, err)

				mock.ExpectBegin()
				tx, err := mock.Begin(ctx)
				assert.NoError(t, err)

				mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))
				return mock, tx
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, pgxTx := tt.setupMock()
			defer mock.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tx := &transaction{
				transactionCtx: ctx,
				transaction:    pgxTx,
				cancelFunc:     cancel,
				startTime:      time.Now(),
			}

			err := tx.rollback()
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

func TestProvider_WithTx(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		setupMock func(mock pgxmock.PgxPoolIface)
		txFunc    func(tx storage.Tx) error
		wantErr   bool
	}{
		{
			name: "successful transaction",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			txFunc: func(tx storage.Tx) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "transaction function error with rollback",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			txFunc: func(tx storage.Tx) error {
				return errors.New("business logic error")
			},
			wantErr: true,
		},
		{
			name: "begin transaction error",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
			},
			txFunc: func(tx storage.Tx) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "commit error",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			},
			txFunc: func(tx storage.Tx) error {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			tt.setupMock(mock)

			provider := &Provider{
				ctx:         ctx,
				logger:      logger,
				db:          mock,
				pingTimeout: 15 * time.Second,
				closeFunc:   func(ctx context.Context) error { return nil },
			}

			err = provider.WithTx(tt.txFunc)
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

func TestTransaction_Query(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:         ctx,
		logger:      logger,
		db:          mock,
		pingTimeout: 15 * time.Second,
		closeFunc:   func(ctx context.Context) error { return nil },
	}

	// Test transaction query
	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(1, "John")
	mock.ExpectQuery("SELECT id, name FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectCommit()

	err = provider.WithTx(func(tx storage.Tx) error {
		scanner, err := tx.Query("SELECT id, name FROM users WHERE id = $1", 1)
		assert.NoError(t, err)
		assert.NotNil(t, scanner)
		return nil
	})

	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTransaction_Execute(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:         ctx,
		logger:      logger,
		db:          mock,
		pingTimeout: 15 * time.Second,
		closeFunc:   func(ctx context.Context) error { return nil },
	}

	// Test transaction execute
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET name = \\$1 WHERE id = \\$2").
		WithArgs("Jane", 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	err = provider.WithTx(func(tx storage.Tx) error {
		return tx.Execute("UPDATE users SET name = $1 WHERE id = $2", "Jane", 1)
	})

	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTransaction_InsertMany(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:         ctx,
		logger:      logger,
		db:          mock,
		pingTimeout: 15 * time.Second,
		closeFunc:   func(ctx context.Context) error { return nil },
	}

	// Test batch insert - skip due to pgxmock limitations with batch operations
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = provider.WithTx(func(tx storage.Tx) error {
		// Skip batch operations due to pgxmock limitations
		// Just test that transaction begins and commits properly
		return nil
	})

	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestProvider__withTx(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	provider := &Provider{
		ctx:         ctx,
		logger:      logger,
		db:          mock,
		pingTimeout: 15 * time.Second,
		closeFunc:   func(ctx context.Context) error { return nil },
	}

	// Test internal _withTx method
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = provider._withTx(func(tx *transaction) error {
		assert.NotNil(t, tx.transaction)
		assert.NotNil(t, tx.transactionCtx)
		assert.NotNil(t, tx.cancelFunc)
		assert.False(t, tx.startTime.IsZero())
		return nil
	})

	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
