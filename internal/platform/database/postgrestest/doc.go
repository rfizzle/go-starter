//go:build integration

// Package postgrestest provides isolated PostgreSQL databases for integration
// tests. It starts a disposable Testcontainers-managed PostgreSQL instance,
// registers all resource cleanup through testing.T, and creates a distinct
// database for every Database or EmptyDatabase call.
package postgrestest
