-- name: CreateSecurityPermissionKey :exec
INSERT INTO security_permission_keys(
    permission_key,
    permission_type,
    permission_key_description,
    created_at
) VALUES($1,$2,$3,$4);

-- name: GetPermissionKeys :many
SELECT permission_key,
    permission_type,
    permission_key_description,
    created_at,
    updated_at
FROM security_permission_keys
WHERE permission_key = ANY($1::varchar[]);

-- name: CreateSecurityRole :one
INSERT INTO security_roles(
    role_uuid,
    role_name,
    created_at
) VALUES ($1,$2,$3) RETURNING role_id;

-- name: CreateSecurityRolePermission :exec
INSERT INTO security_role_permissions(
    role_id,
    permission_key,
    permission_values,
    permission_bits_value,
    row_version,
    created_at,
    updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7);

-- name: GetSecurityRolePermissions :many
select role_id,
    permission_key,
    permission_values,
    permission_bits_value,
    row_version,
    created_at,
    updated_at
FROM security_role_permissions
WHERE role_id = $1;
