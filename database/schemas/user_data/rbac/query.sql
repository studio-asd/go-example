-- name: CreatePermissions :one
INSERT INTO security_permissions(
    permission_name,
    permission_type,
    permission_value,
    created_at
) VALUES ($1,$2,$3,$4) RETURNING permission_id;

-- name: GetPermissions :many
SELECT permission_id,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_id = ANY($1::bigint[]);

-- name: CreateSecurityRole :one
INSERT INTO security_roles(
    role_name,
    created_at
) VALUES ($1,$2) RETURNING role_id;

-- name: CreateSecurityRolePermission :exec
INSERT INTO security_role_permissions(
    role_id,
    permission_id,
    created_at
) VALUES ($1,$2,$3);