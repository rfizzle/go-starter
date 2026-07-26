//go:build integration

// Package postgrestest provides isolated PostgreSQL databases for integration
// tests. Each test binary starts one disposable Testcontainers-managed
// PostgreSQL instance through RunMain and Shared, creates a distinct database
// for every Database or EmptyDatabase call, and registers per-database cleanup
// through testing.T.
package postgrestest
