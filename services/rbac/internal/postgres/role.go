package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/services/rbac"
)

type CreateRole struct {
	RoleExternalID string
	RoleName       string
	CreatedAt      time.Time
	PermissionIDs  []int64
}

func (q *Queries) CreateRole(ctx context.Context, role CreateRole) (int64, error) {
	var (
		roleID int64
		err    error
	)

	errTransact := q.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *Queries) error {
		roleID, err = q.CreateSecurityRole(ctx, CreateSecurityRoleParams{
			RoleExternalID: role.RoleExternalID,
			RoleName:       role.RoleName,
			CreatedAt:      role.CreatedAt,
		})
		if err != nil {
			return err
		}
		// Create all the params based on the role id.
		params := make([]any, len(role.PermissionIDs)*3)
		for i, permID := range role.PermissionIDs {
			idx := i * 3
			params[idx+0] = roleID
			params[idx+1] = permID
			params[idx+2] = role.CreatedAt
		}

		return q.db.BulkInsert(
			ctx,
			"security_role_permissions",
			[]string{
				"role_id",
				"permission_id",
				"created_at",
			},
			params,
			postgres.OnConflictDoNothing,
		)
	})
	if errTransact != nil {
		return 0, errTransact
	}
	return roleID, nil
}

type RolePermissions struct {
	RoleID         int64
	RoleExternalID string
	RoleName       string
	Permissions    []Permission
	CreatedAt      time.Time
}

type Permission struct {
	PermissionID         int64
	PermissionExternalID string
	PermissionName       string
	PermissionKey        string
	PermissionValue      string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (q *Queries) GetRolePermissions(ctx context.Context, roleID int64) (RolePermissions, error) {
	result := RolePermissions{
		RoleID: roleID,
	}
	rolePerms, err := q.GetSecurityRolePermissions(ctx, roleID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return result, rbac.ErrRoleNotFound
		}
		return result, err
	}
	if len(rolePerms) == 0 {
		return result, rbac.ErrRoleNotFound
	}

	result.Permissions = make([]Permission, len(rolePerms))
	result.RoleName = rolePerms[0].RoleName
	for idx, perm := range rolePerms {
		result.Permissions[idx] = Permission{
			PermissionID:         perm.PermissionID,
			PermissionExternalID: perm.PermissionExternalID,
			PermissionName:       perm.PermissionName,
			PermissionKey:        perm.PermissionKey,
			PermissionValue:      perm.PermissionValue,
		}
	}
	return result, nil
}
