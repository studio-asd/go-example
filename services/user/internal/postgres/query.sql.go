// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package postgres

import (
	"context"
	"time"
)

const createUser = `-- name: CreateUser :one
INSERT INTO user_data.users (
	external_id,
	user_email,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4) RETURNING user_id
`

type CreateUserParams struct {
	ExternalID string
	UserEmail  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ExternalID,
		arg.UserEmail,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var user_id int64
	err := row.Scan(&user_id)
	return user_id, err
}

const createUserPII = `-- name: CreateUserPII :exec
INSERT INTO user_data.users_pii(
	user_id,
	phone_number,
	identity_number,
	identity_type,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4,$5,$6)
`

type CreateUserPIIParams struct {
	UserID         int64
	PhoneNumber    string
	IdentityNumber string
	IdentityType   int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (q *Queries) CreateUserPII(ctx context.Context, arg CreateUserPIIParams) error {
	_, err := q.db.Exec(ctx, createUserPII,
		arg.UserID,
		arg.PhoneNumber,
		arg.IdentityNumber,
		arg.IdentityType,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const createUserSession = `-- name: CreateUserSession :exec
INSERT INTO user_data.user_sessions(
	user_id,
	random_number,
	created_time,
	session_metadata,
	expired_at
) VALUES($1,$2,$3,$4,$5)
`

type CreateUserSessionParams struct {
	UserID          int64
	RandomNumber    int32
	CreatedTime     int64
	SessionMetadata []byte
	ExpiredAt       time.Time
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) error {
	_, err := q.db.Exec(ctx, createUserSession,
		arg.UserID,
		arg.RandomNumber,
		arg.CreatedTime,
		arg.SessionMetadata,
		arg.ExpiredAt,
	)
	return err
}

const getUserSession = `-- name: GetUserSession :one
SELECT user_id, random_number, created_time, session_metadata, expired_at
FROM user_data.user_sessions
WHERE user_id = $1
	AND random_number = $2
	AND created_time = $3
`

type GetUserSessionParams struct {
	UserID       int64
	RandomNumber int32
	CreatedTime  int64
}

func (q *Queries) GetUserSession(ctx context.Context, arg GetUserSessionParams) (UserDataUserSession, error) {
	row := q.db.QueryRow(ctx, getUserSession, arg.UserID, arg.RandomNumber, arg.CreatedTime)
	var i UserDataUserSession
	err := row.Scan(
		&i.UserID,
		&i.RandomNumber,
		&i.CreatedTime,
		&i.SessionMetadata,
		&i.ExpiredAt,
	)
	return i, err
}
