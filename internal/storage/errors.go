package storage

import "errors"

// ErrNotFound is the shared persistence outcome consumed by controllers and
// commands. Repositories translate pgx.ErrNoRows to this sentinel so callers
// can handle absence without depending on the PostgreSQL driver.
var ErrNotFound = errors.New("resource not found")
