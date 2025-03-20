-- name: CreateUser :one
INSERT INTO user_data.users (
	external_id,
	user_email,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4) RETURNING user_id;

-- name: CreateUserPII :exec
INSERT INTO user_data.users_pii(
	user_id,
	phone_number,
	identity_number,
	identity_type,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4,$5,$6);

-- name: CreateUserSession :exec
INSERT INTO user_data.user_sessions(
	user_id,
	random_number,
	created_time,
	created_from_ip,
	created_from_loc,
	created_from_user_agent,
	session_metadata,
	expired_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8);

-- name: GetUserSession :one
SELECT *
FROM user_data.user_sessions
WHERE user_id = $1
	AND random_number = $2
	AND created_time = $3;

-- name: CreateUserSecret :one
INSERT INTO user_data.user_secrets(
	external_id,
    user_id,
    secret_key,
    secret_type,
    current_secret_version,
    created_at
) VALUES($1,$2,$3,$4,$5,$6) RETURNING secret_id;

-- name: CreateUserSecretVersion :exec
INSERT INTO user_data.user_secret_versions(
    secret_id,
    secret_version,
    secret_value,
    created_at
) VALUES($1,$2,$3,$4);

-- name: GetUserSecret :one
SELECT *
FROM user_data.user_secrets
WHERE user_id = $1
	AND secret_key = $2
	AND secret_type = $3;

-- name: GetUserSecretByType :many
SELECT *
FROM user_data.user_secrets
WHERE user_id = $1
	AND secret_type = $2;

-- name: GetUserSecretValue :one
SELECT us.secret_id,
	us.external_id,
	us.user_id,
	us.secret_key,
	us.secret_type,
	us.current_secret_version,
	us.created_at,
	-- The updated_at is the same with the new version created_at so we don't
	-- have to retrieve more information from usv.
	us.updated_at,
	usv.secret_value
FROM user_data.user_secrets us,
	user_data.user_secret_versions usv
WHERE us.user_id = $1
	AND us.secret_key = $2
	AND us.secret_type = $3
	AND us.current_secret_version = usv.secret_version
	AND us.secret_id = usv.secret_id;