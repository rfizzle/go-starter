package postgres

import "embed"

// Migrations holds the embedded, forward-only SQL migrations applied by
// Provider.Migrate at startup. It is the same directory sqlc reads its schema
// from (see sqlc.yaml), so the generated types and the live schema share a
// single source of truth.
//
//go:embed migrations/*.sql
var Migrations embed.FS
