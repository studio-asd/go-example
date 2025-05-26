package postgres

import (
	"context"

	"github.com/studio-asd/pkg/postgres"
)

func (q *Queries) CreatePermissions(ctx context.Context, permissions ...SecurityPermission) ([]int64, error) {
	columns := []string{
		"permission_uuid",
		"permission_name",
		"permission_type",
		"permission_key",
		"permission_value",
		"created_at",
	}
	// returningColumns is used to retrieve the generated permission ids as we need to expose the ids
	// to other functions so it can be referenced.
	returningColumns := []string{
		"permission_id",
	}

	// Create all the params based on the permissions.
	params := make([]any, len(permissions)*6)
	for i, perm := range permissions {
		idx := i * 6
		params[idx+0] = perm.PermissionUuid
		params[idx+1] = perm.PermissionName
		params[idx+2] = perm.PermissionType
		params[idx+3] = perm.PermissionKey
		params[idx+4] = perm.PermissionValue
		params[idx+5] = perm.CreatedAt
	}

	var permissionIDs []int64
	err := q.db.WithMetrics(ctx, "createPermissions", func(ctx context.Context, pg *postgres.Postgres) error {
		return pg.BulkInsertReturning(
			ctx,
			"security_permissions",
			columns,
			params,
			"",
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
