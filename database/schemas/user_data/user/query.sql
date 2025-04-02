-- name: CreateUser :one
INSERT INTO users (
	external_id,
	created_at,
	updated_at
) VALUES($1,$2,$3) RETURNING user_id;

-- name: GetUserByExternalID :one
SELECT usr.user_id,
	usr.external_id,
	usr.created_at,
	usr.updated_at,
	upi.email
FROM users usr,
    user_pii upi
WHERE usr.external_id = $1
    AND usr.user_id = upi.user_id;

-- name: GetUserByEmail :one
SELECT usr.user_id,
usr.external_id,
usr.created_at,
usr.updated_at,
upi.email
FROM users usr,
    user_pii upi
WHERE upi.email = $1
    AND usr.user_id = upi.user_id;

-- name: CreateUserPII :exec
INSERT INTO user_pii (
	user_id,
	email,
	phone_number,
	identity_number,
	identity_type,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4,$5,$6,$7);

-- name: GetUserPII :one
SELECT user_id,
	email,
	phone_number,
	identity_number,
	identity_type,
	created_at,
	updated_at
FROM user_pii
WHERE user_id = $1;

-- name: CreateUserSession :exec
INSERT INTO user_sessions(
	session_id,
	session_type,
	user_id,
	random_id,
	created_from_ip,
	created_from_loc,
	created_from_user_agent,
	session_metadata,
	created_at,
	expired_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);

-- name: GetUserSession :one
SELECT us.session_id,
	us.previous_sesision_id,
	us.session_type,
	us.user_id,
	up.email,
	us.random_id,
	us.created_from_ip,
	us.created_from_loc,
	us.created_from_user_agent,
	us.session_metadata,
	us.created_at,
	us.expired_at
FROM user_sessions us
	LEFT JOIN user_pii up ON
		us.user_id IS NOT NULL AND
		us.user_id = up.user_id
WHERE us.session_id = $1;

-- name: CreateUserSecret :one
INSERT INTO user_secrets(
	external_id,
    user_id,
    secret_key,
    secret_type,
    current_secret_version,
    created_at
) VALUES($1,$2,$3,$4,$5,$6) RETURNING secret_id;

-- name: CreateUserSecretVersion :exec
INSERT INTO user_secret_versions(
    secret_id,
    secret_version,
    secret_value,
	secret_salt,
    created_at
) VALUES($1,$2,$3,$4,$5);

-- name: GetUserSecret :one
SELECT secret_id,
	external_id,
	user_id,
	secret_key,
	secret_type,
	current_secret_version,
	created_at,
	updated_at
FROM user_secrets
WHERE user_id = $1
	AND secret_key = $2
	AND secret_type = $3;

-- name: GetUserSecretByType :many
SELECT secret_id,
	external_id,
	user_id,
	secret_key,
	secret_type,
	current_secret_version,
	created_at,
	updated_at
FROM user_secrets
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
	usv.secret_value,
	usv.secret_salt
FROM user_secrets us,
	user_secret_versions usv
WHERE us.user_id = $1
	AND us.secret_key = $2
	AND us.secret_type = $3
	AND us.current_secret_version = usv.secret_version
	AND us.secret_id = usv.secret_id;

-- name: GetUserSecretByExternalID :one
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
	usv.secret_value,
	usv.secret_salt
FROM user_secrets us,
	user_secret_versions usv
WHERE us.external_id = $1
	AND us.current_secret_version = usv.secret_version
	AND us.secret_id = usv.secret_id;
