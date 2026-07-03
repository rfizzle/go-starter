package storage

import (
	"context"

	"go.uber.org/zap"
)

// RowScanner provides methods for scanning query results into Go types.
// Renamed from Scanner to avoid naming conflicts with other scanner interfaces.
type RowScanner interface {
	// ScanOne scans the first row from the result set into dst.
	// Returns an error if no rows are available or scanning fails.
	ScanOne(dst interface{}) error
	
	// ScanAll scans all rows from the result set into dst.
	// dst must be a pointer to a slice of structs.
	ScanAll(dst interface{}) error
}

// Client defines the core database operations and lifecycle management.
// It provides a complete interface for database interactions including
// queries, transactions, and connection management.
type Client interface {
	// Query executes a SQL query and returns a RowScanner for reading results.
	// The context can be used to cancel the operation.
	Query(ctx context.Context, sql string, args ...interface{}) (RowScanner, error)
	
	// Execute runs a SQL statement (INSERT, UPDATE, DELETE) without returning results.
	// The context can be used to cancel the operation.
	Execute(ctx context.Context, sql string, args ...interface{}) error

	// BeginTx starts a new database transaction and returns a Transaction interface.
	// The context can be used to set timeouts or cancel the operation.
	BeginTx(ctx context.Context) (Transaction, error)
	
	// WithTx executes a function within a database transaction.
	// If fn returns an error, the transaction is rolled back.
	// Otherwise, the transaction is committed.
	WithTx(ctx context.Context, fn func(tx Transaction) error) error

	// Ping verifies the database connection is alive.
	// Returns an error if the connection is not available.
	Ping(ctx context.Context) error
	
	// Close closes the database connection and releases resources.
	// Should be called when the client is no longer needed.
	Close() error

	// Name returns the name identifier of the database provider.
	Name() string
	
	// Logger returns the logger instance used by this client.
	Logger() *zap.Logger
}

// Transaction represents a database transaction with ACID properties.
// All operations within a transaction are executed atomically and can be
// committed or rolled back as a single unit.
type Transaction interface {
	// Query executes a SQL query within the transaction context.
	// Results are part of the transaction and will be committed or rolled back.
	Query(ctx context.Context, sql string, args ...interface{}) (RowScanner, error)
	
	// Execute runs a SQL statement within the transaction context.
	// Changes are part of the transaction and will be committed or rolled back.
	Execute(ctx context.Context, sql string, args ...interface{}) error

	// InsertMany executes multiple INSERT statements atomically within the transaction.
	// queries is a map of SQL statements to their respective parameter lists.
	InsertMany(ctx context.Context, queries map[string][]interface{}) error
	
	// Commit permanently saves all changes made within this transaction.
	// After commit, the transaction cannot be used for further operations.
	Commit() error
	
	// Rollback discards all changes made within this transaction.
	// After rollback, the transaction cannot be used for further operations.
	Rollback() error
}
