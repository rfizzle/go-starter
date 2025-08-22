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

func (provider *Provider) begin() (*transaction, error) {
	ctx, cancelFunc := context.WithTimeout(provider.ctx, provider.pingTimeout)

	tx, err := provider.db.Begin(ctx)
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

func (tx *transaction) commit() error {
	defer tx.cancelFunc()
	return tx.transaction.Commit(tx.transactionCtx)
}

func (tx *transaction) rollback() error {
	defer tx.cancelFunc()
	return tx.transaction.Rollback(tx.transactionCtx)
}

// WithTx creates a function with core.Transaction and core.Provider to be called throughout the app
func (provider *Provider) WithTx(fn func(tx storage.Tx) error) error {
	startTime := time.Now()
	tx, err := provider.begin()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		if rollbackErr := tx.rollback(); rollbackErr != nil {
			return err // TODO: something important here
		}
		return err
	}
	err = tx.commit()
	stopTime := time.Now()
	provider.logger.Debug(fmt.Sprintf(`transaction completed in %v`, stopTime.Sub(startTime)))
	return err
}

// _withTx creates a function with transaction and postgresqlStore to be used internally within the package
func (provider *Provider) _withTx(fn func(tx *transaction) error) error {
	startTime := time.Now()
	tx, err := provider.begin()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		if rollbackErr := tx.rollback(); rollbackErr != nil {
			return err // TODO: something important here
		}
		return err
	}
	err = tx.commit()
	stopTime := time.Now()
	provider.logger.Debug(fmt.Sprintf(`transaction completed in %v`, stopTime.Sub(startTime)))
	return err
}

func (tx *transaction) Query(sql string, args ...interface{}) (storage.Scanner, error) {
	rows, err := tx._query(sql, args...)
	if err != nil {
		return nil, err
	}

	return &internalScanner{rows: rows}, nil
}

func (tx *transaction) Execute(sql string, args ...interface{}) error {
	return tx._exec(sql, args...)
}

func (tx *transaction) InsertMany(queries map[string][]interface{}) error {
	return tx._batch(queries)
}

func (tx *transaction) _query(sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := tx.transaction.Query(tx.transactionCtx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// _exec is a generic transaction execution function for pgx
func (tx *transaction) _exec(sql string, args ...interface{}) error {
	_, err := tx.transaction.Exec(tx.transactionCtx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}

// _exec is a generic batch execution function with a transaction for pgx
func (tx *transaction) _batch(queries map[string][]interface{}) error {
	// Create batch
	batch := &pgx.Batch{}

	// Enqueue queries to batch
	for q, args := range queries {
		batch.Queue(q, args...)
	}

	// Send batch
	results := tx.transaction.SendBatch(tx.transactionCtx, batch)

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
