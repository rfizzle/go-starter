package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/rfizzle/go-starter/internal/storage"

	"github.com/jackc/pgx/v5"
)

type transaction struct {
	transactionCtx context.Context
	transaction    pgx.Tx
	cancelFunc     context.CancelFunc
	startTime      time.Time
}

func (client *Client) begin(ctx context.Context) (*transaction, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, client.pingTimeout)

	tx, err := client.db.Begin(ctx)
	if err != nil {
		cancelFunc()
		return nil, err
	}

	return &transaction{
		transactionCtx: ctx,
		transaction:    tx,
		cancelFunc:     cancelFunc,
		startTime:      time.Now(),
	}, nil
}

func (tx *transaction) Commit() error {
	defer tx.cancelFunc()
	return tx.transaction.Commit(tx.transactionCtx)
}

func (tx *transaction) Rollback() error {
	defer tx.cancelFunc()
	return tx.transaction.Rollback(tx.transactionCtx)
}

// BeginTx starts a new database transaction
func (client *Client) BeginTx(ctx context.Context) (storage.Transaction, error) {
	return client.begin(ctx)
}

// WithTx creates a function with storage.Transaction to be called throughout the app
func (client *Client) WithTx(ctx context.Context, fn func(tx storage.Transaction) error) error {
	startTime := time.Now()
	tx, err := client.begin(ctx)
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Return a wrapped error that includes both the original error and rollback failure
			return fmt.Errorf("transaction failed: %w (rollback also failed: %v)", err, rollbackErr)
		}
		return err
	}
	err = tx.Commit()
	stopTime := time.Now()
	client.logger.Debug(fmt.Sprintf(`transaction completed in %v`, stopTime.Sub(startTime)))
	return err
}


func (tx *transaction) Query(ctx context.Context, sql string, args ...interface{}) (storage.RowScanner, error) {
	rows, err := tx._query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return &internalScanner{rows: rows}, nil
}

func (tx *transaction) Execute(ctx context.Context, sql string, args ...interface{}) error {
	return tx._exec(ctx, sql, args...)
}

func (tx *transaction) InsertMany(ctx context.Context, queries map[string][]interface{}) error {
	return tx._batch(ctx, queries)
}

func (tx *transaction) _query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := tx.transaction.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// _exec is a generic transaction execution function for pgx
func (tx *transaction) _exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := tx.transaction.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}

// _batch is a generic batch execution function with a transaction for pgx
func (tx *transaction) _batch(ctx context.Context, queries map[string][]interface{}) error {
	// Create batch
	batch := &pgx.Batch{}

	// Enqueue queries to batch
	for q, args := range queries {
		batch.Queue(q, args...)
	}

	// Send batch
	results := tx.transaction.SendBatch(ctx, batch)

	// Get batch results
	for i := 0; i < len(queries); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	// Close the batch connection
	err := results.Close()
	if err != nil {
		return err
	}

	return nil
}
