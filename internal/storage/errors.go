package storage

import "errors"

// Common storage errors that may be returned by storage implementations.
// These provide a consistent error interface across different database backends.
var (
	// ErrNotFound is returned when a requested resource is not found in the database.
	// This typically occurs during SELECT, UPDATE, or DELETE operations when
	// no matching records exist.
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicate is returned when attempting to create a resource that already exists.
	// This usually occurs during INSERT operations when a unique constraint is violated.
	ErrDuplicate = errors.New("resource already exists")

	// ErrConnection is returned when there's a database connection error.
	// This can happen due to network issues, authentication failures, or server unavailability.
	ErrConnection = errors.New("database connection error")

	// ErrTransaction is returned when a transaction operation fails.
	// This includes failures to begin, commit, or rollback a transaction.
	ErrTransaction = errors.New("transaction error")

	// ErrInvalidQuery is returned when a query is malformed or invalid.
	// This occurs when SQL syntax is incorrect or references non-existent tables/columns.
	ErrInvalidQuery = errors.New("invalid query")

	// ErrTimeout is returned when a database operation times out.
	// This happens when operations take longer than the configured timeout duration.
	ErrTimeout = errors.New("operation timeout")
)