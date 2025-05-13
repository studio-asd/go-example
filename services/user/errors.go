package user

import "errors"

var (
	// User errors.
	ErrUserNotFound        = errors.New("user: not found")
	ErrUserSessionNotFound = errors.New("user: session not found")
	// Password errors.
	ErrPasswordInvalid = errors.New("user: invalid password")
	// Session errors.
	ErrSessionExpired          = errors.New("session: session expired")
	ErrSessionUserIDEmpty      = errors.New("session: user ID is empty")
	ErrSessionRandomIDEmpty    = errors.New("session: random ID is empty")
	ErrSessionCreatedAtInvalid = errors.New("session: created at timestamp invalid")
	ErrSessionCreatedAtTooOld  = errors.New("session: created at timestamp is too old")
)
