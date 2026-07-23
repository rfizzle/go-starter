// Package database defines technical database lifecycle contracts shared by
// PostgreSQL infrastructure. Business-facing models and errors live in domain.
package database

// MigrationStatus represents the state of database migrations.
// It indicates how many of the available migration files have been
// successfully applied to the database.
type MigrationStatus int

// Migration status constants using idiomatic Go naming.
// These constants represent the different states of database migration completion.
const (
	// StatusNone indicates that no migrations have been applied to the database.
	// This typically occurs on a fresh database installation.
	StatusNone MigrationStatus = 0

	// StatusPartial indicates that some, but not all, migrations have been applied.
	// This may occur when migrations are added over time or if a migration failed.
	StatusPartial MigrationStatus = 1

	// StatusComplete indicates that all available migrations have been successfully applied.
	// The database schema is current with all migration files.
	StatusComplete MigrationStatus = 2
)
