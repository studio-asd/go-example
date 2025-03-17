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
	created_from_macaddr,
	created_from_loc,
	created_from_user_agent,
	session_metadata,
	expired_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9);

-- name: GetUserSession :one
SELECT *
FROM user_data.user_sessions
WHERE user_id = $1
	AND random_number = $2
	AND created_time = $3;
