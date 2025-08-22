package storage

// MigrationStatus is an internal type for migration state
type MigrationStatus int

// Constant variables for all possible migration states
const (
	NoMigrationsCompleted   MigrationStatus = 0
	SomeMigrationsCompleted MigrationStatus = 1
	AllMigrationsCompleted  MigrationStatus = 2
)
