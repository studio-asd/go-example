-- name: GetPermissions :many
SELECT permission_id,
    permission_uuid,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_id = ANY($1::bigint[]);

-- name: GetPermissionsByUUID :many
SELECT permission_id,
    permission_uuid,
    permission_name,
    permission_type,
    permission_value,
    created_at,
    updated_at
FROM security_permissions
WHERE permission_uuid = ANY($1::uuid[]);

-- name: CreateSecurityRole :one
INSERT INTO security_roles(
    role_uuid,
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
    sr.role_name,
    sp.permission_id,
    sp.permission_uuid,
    sp.permission_name,
    sp.permission_type,
    sp.permission_key,
    sp.permission_value
FROM security_roles sr,
    security_permissions sp,
    security_role_permissions srp
WHERE sr.role_id = $1
    AND sr.role_id = srp.role_id
    AND srp.permission_id = sp.permission_id
ORDER by srp.role_id, srp.permission_id;
