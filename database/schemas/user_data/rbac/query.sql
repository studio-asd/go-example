-- name: GetPermissions :many
SELECT permission_id,
    permission_external_id,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_id = ANY($1::bigint[]);
--
-- name: GetPermissionsByExternalIDs :many
SELECT permission_id,
    permission_external_id,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_external_id = ANY($1::varchar[]);

-- name: CreateSecurityRole :one
INSERT INTO security_roles(
    role_external_id,
    role_name,
    created_at
) VALUES ($1,$2,$3) RETURNING role_id;

-- name: CreateSecurityRolePermission :exec
INSERT INTO security_role_permissions(
    role_id,
    permission_id,
    created_at
) VALUES ($1,$2,$3);

-- name: GetSecurityRolePermissions :many
SELECT srp.role_id,
    sp.permission_id,
    sp.permission_external_id,
    sp.permission_name,
    sp.permission_type,
    sp.permission_value
FROM security_permissions sp
JOIN security_role_permissions srp
    ON sp.permission_id = srp.permission_id
WHERE srp.role_id = $1;
