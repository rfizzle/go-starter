package postgres

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/rfizzle/go-starter/internal/storage"
)

type migration struct {
	MigrationID string    `json:"migration_id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Migrate executes pending database migrations from the provided embedded filesystem.
// It reads .sql files from migrationFs, compares them with executed migrations,
// and runs any new migrations atomically within a transaction.
// Returns an error if any migration fails.
func (provider *Provider) Migrate(migrationFs embed.FS) error {
	// Setup context
	ctx := context.Background()

	// Add migration table if not exists
	err := provider._addMigrationTable(ctx)
	if err != nil {
		provider.logger.Error("issue adding migration table")
		return err
	}

	// Get migration current state
	dbMigrations, err := provider._getDbMigrations(ctx)
	if err != nil {
		provider.logger.Error("issue getting database migrations")
		return err
	}

	fileMigrations, err := provider._getMigrationsFromFs(ctx, migrationFs)
	if err != nil {
		provider.logger.Error("issue getting migrations from virtual filesystem")
		return err
	}

	// Find the new migrations
	newMigrations := filterExecutedMigrations(fileMigrations, dbMigrations)

	// Return if there are no new migrations
	if len(newMigrations) == 0 {
		return nil
	}

	// Execute migrations inside a transaction
	err = provider.WithTx(ctx, func(tx storage.Transaction) error {
		for _, v := range newMigrations {
			// Read sql file
			data, err := fs.ReadFile(migrationFs, v)
			if err != nil {
				provider.logger.Error("issue reading migration file from virtual filesystem")
				return err
			}

			// Execute sql file
			err = tx.Execute(ctx, string(data))
			if err != nil {
				provider.logger.Error("issue running migration")
				return err
			}

			// Add the new migration to the database
			err = tx.Execute(ctx, `INSERT into migrations(id) VALUES ($1::text);`, v)
			if err != nil {
				provider.logger.Error("issue adding executed migration to row")
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// MigrationStatus returns the current status of database migrations.
// It compares migration files in migrationFs with executed migrations in the database
// to determine if all migrations are applied, partially applied, or none are applied.
func (provider *Provider) MigrationStatus(ctx context.Context, migrationFs embed.FS) (storage.MigrationStatus, error) {
	err := provider._addMigrationTable(ctx)
	if err != nil {
		return storage.StatusNone, err
	}

	fileMigrations, err := provider._getMigrationsFromFs(ctx, migrationFs)
	if err != nil {
		return storage.StatusNone, err
	}

	dbMigrations, err := provider._getDbMigrations(ctx)
	if err != nil {
		return storage.StatusNone, err
	}

	if len(dbMigrations) == len(fileMigrations) {
		return storage.StatusComplete, nil
	} else if len(dbMigrations) == 0 {
		return storage.StatusNone, nil
	}
	return storage.StatusPartial, nil
}

func (provider *Provider) _addMigrationTable(ctx context.Context) error {
	err := provider.WithTx(ctx, func(tx storage.Transaction) error {
		// Setup UUID extension
		err := tx.Execute(ctx,
			`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		)
		if err != nil {
			provider.logger.Error("issue creating uuid-ossp extension")
			return err
		}

		// Setup migrations table
		err = tx.Execute(ctx,
			`CREATE TABLE IF NOT EXISTS migrations (id text NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), PRIMARY KEY (id));`,
		)
		if err != nil {
			provider.logger.Error("issue creating migrations table")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (provider *Provider) _getMigrationsFromFs(ctx context.Context, migrationFs embed.FS) ([]string, error) {
	migrationNames := make([]string, 0)

	// Read all migration files
	err := fs.WalkDir(migrationFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".sql" {
			migrationNames = append(migrationNames, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return migrationNames, nil
}

// _getDbMigrations returns the successfully executed migrations
func (provider *Provider) _getDbMigrations(ctx context.Context) ([]*migration, error) {
	var migrations []*migration
	values, err := provider.Query(ctx, `SELECT id, created_at FROM migrations`)
	if err != nil {
		return nil, err
	}

	err = values.ScanAll(&migrations)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

// filterExecutedMigrations takes the migrations from the files and the database and returns just the unexecuted ones
func filterExecutedMigrations(fileMigrations []string, dbMigrations []*migration) []string {
	m := make([]string, 0)

	for _, v := range fileMigrations {
		found := false
		for _, v2 := range dbMigrations {
			if v == v2.MigrationID {
				found = true
			}
		}

		if !found {
			m = append(m, v)
		}
	}

	return m
}
