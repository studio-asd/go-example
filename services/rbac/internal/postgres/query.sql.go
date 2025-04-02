// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package postgres

import (
	"context"
	"database/sql"
	"time"
)

const createPermissions = `-- name: CreatePermissions :one
INSERT INTO security_permissions(
    permission_name,
    permission_type,
    permission_value,
    created_at
) VALUES ($1,$2,$3,$4) RETURNING permission_id
`

type CreatePermissionsParams struct {
	PermissionName  string
	PermissionType  int32
	PermissionValue string
	CreatedAt       time.Time
}

func (q *Queries) CreatePermissions(ctx context.Context, arg CreatePermissionsParams) (int64, error) {
	row := q.db.QueryRow(ctx, createPermissions,
		arg.PermissionName,
		arg.PermissionType,
		arg.PermissionValue,
		arg.CreatedAt,
	)
	var permission_id int64
	err := row.Scan(&permission_id)
	return permission_id, err
}

const createSecurityRole = `-- name: CreateSecurityRole :one
INSERT INTO security_roles(
    role_name,
    created_at
) VALUES ($1,$2) RETURNING role_id
`

type CreateSecurityRoleParams struct {
	RoleName  string
	CreatedAt time.Time
}

func (q *Queries) CreateSecurityRole(ctx context.Context, arg CreateSecurityRoleParams) (int64, error) {
	row := q.db.QueryRow(ctx, createSecurityRole, arg.RoleName, arg.CreatedAt)
	var role_id int64
	err := row.Scan(&role_id)
	return role_id, err
}

const createSecurityRolePermission = `-- name: CreateSecurityRolePermission :exec
INSERT INTO security_role_permissions(
    role_id,
    permission_id,
    created_at
) VALUES ($1,$2,$3)
`

type CreateSecurityRolePermissionParams struct {
	RoleID       int64
	PermissionID int64
	CreatedAt    time.Time
}

func (q *Queries) CreateSecurityRolePermission(ctx context.Context, arg CreateSecurityRolePermissionParams) error {
	_, err := q.db.Exec(ctx, createSecurityRolePermission, arg.RoleID, arg.PermissionID, arg.CreatedAt)
	return err
}

const getPermissions = `-- name: GetPermissions :many
SELECT permission_id,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_id = ANY($1::bigint[])
`

type GetPermissionsRow struct {
	PermissionID    int64
	PermissionName  string
	PermissionType  int32
	PermissionValue string
	CreatedAt       time.Time
	UpdatedAt       sql.NullTime
}

func (q *Queries) GetPermissions(ctx context.Context, dollar_1 []int64) ([]GetPermissionsRow, error) {
	rows, err := q.db.Query(ctx, getPermissions, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPermissionsRow
	for rows.Next() {
		var i GetPermissionsRow
		if err := rows.Scan(
			&i.PermissionID,
			&i.PermissionName,
			&i.PermissionType,
			&i.PermissionValue,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
