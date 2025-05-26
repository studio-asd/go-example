package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/services/rbac"
)

type CreateRole struct {
	RoleUUID      uuid.UUID
	RoleName      string
	CreatedAt     time.Time
	PermissionIDs []int64
}

func (q *Queries) CreateRole(ctx context.Context, role CreateRole) (int64, error) {
	var (
		roleID int64
		err    error
	)

	errTransact := q.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *Queries) error {
		roleID, err = q.CreateSecurityRole(ctx, CreateSecurityRoleParams{
			RoleUuid:  role.RoleUUID,
			RoleName:  role.RoleName,
			CreatedAt: role.CreatedAt,
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
			"",
		)
	})
	if errTransact != nil {
		return 0, errTransact
	}
	return roleID, nil
}

type RolePermissions struct {
	RoleID      int64
	RoleUUID    uuid.UUID
	RoleName    string
	Permissions []Permission
	CreatedAt   time.Time
}

type Permission struct {
	PermissionID    int64
	PermissionUUID  uuid.UUID
	PermissionName  string
	PermissionType  string
	PermissionKey   string
	PermissionValue string
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
			PermissionID:    perm.PermissionID,
			PermissionUUID:  perm.PermissionUuid,
			PermissionName:  perm.PermissionName,
			PermissionType:  perm.PermissionType,
			PermissionKey:   perm.PermissionKey,
			PermissionValue: perm.PermissionValue,
		}
	}
	return result, nil
}
