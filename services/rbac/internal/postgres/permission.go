package postgres

import (
	"context"

	"github.com/studio-asd/pkg/postgres"
)

func (q *Queries) CreatePermissions(ctx context.Context, permissions []SecurityPermission) ([]int64, error) {
	columns := []string{
		"permission_external_id",
		"permission_name",
		"permission_type",
	}
	// returningColumns is used to retrieve the generated permission ids as we need to expose the ids
	// to other functions so it can be referenced.
	returningColumns := []string{
		"permission_id",
	}

	var permissionIDs []int64
	err := q.db.WithMetrics(ctx, "createPermissions", func(ctx context.Context, pg *postgres.Postgres) error {
		return pg.BulkInsertReturning(
			ctx,
			"security_permissions",
			columns,
			nil,
			postgres.OnConflictDoNothing,
			returningColumns,
			func(rc *postgres.RowsCompat) error {
				var id int64
				if err := rc.Scan(&id); err != nil {
					return err
				}
				permissionIDs = append(permissionIDs, id)
				return nil
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return permissionIDs, nil
}
