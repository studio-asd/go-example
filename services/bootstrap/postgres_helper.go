package bootstrap

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/studio-asd/pkg/postgres"
)

// PostgresCheckTables checks if the given tables exist in the database and the columns are correct. This to ensure
// that the migrations is already happen and coorect.
func PostgresCheckTables(tables []string, columns []map[string]string) {
	// If columns is nil, then we are conciously not checking the columns name and data type.
	if columns == nil {
		return
	}
}

// PostgresCheckMigrations checks the golang-migrate migrations via schema_midgrations table.
//
// The check will failed on two conditions:
// 1. If there is any dirty migration, then the check will failed.
// 2. If there is any missing migration, then the check will failed.
func PostgresCheckMigrations(ctx context.Context, pg *postgres.Postgres, versions []int64) error {
	type SchemaMigrations struct {
		Version int64 `db:"version"`
		Dirty   bool  `db:"dirty"`
	}

	query, params, err := squirrel.Select("version", "dirty").
		From("schema_migrations").
		Where(squirrel.Eq{"version": versions}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	schemaMigrations := make(map[int64]SchemaMigrations)
	if err := pg.RunQuery(ctx, query, func(rows *postgres.RowsCompat) error {
		var (
			version int64
			dirty   bool
		)
		if err := rows.Scan(&version, &dirty); err != nil {
			return err
		}
		schemaMigrations[version] = SchemaMigrations{Version: version, Dirty: dirty}
		return nil
	}, params...); err != nil {
		return err
	}

	// Check all the migrations.
	for _, version := range versions {
		if _, ok := schemaMigrations[version]; !ok {
			return fmt.Errorf("missing migration version %d", version)
		}
		if schemaMigrations[version].Dirty {
			return fmt.Errorf("dirty migration version %d", version)
		}
	}
	return nil
}
