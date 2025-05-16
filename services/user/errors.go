package user

import "errors"

var (
	// User errors.
	ErrUserNotFound        = errors.New("user: not found")
	ErrUserSessionNotFound = errors.New("user: session not found")
	// Password errors.
	ErrPasswordInvalid   = errors.New("user: invalid password")
	ErrPasswordSaltEmpty = errors.New("user: password salt is empty")
	ErrPasswordTooShort  = errors.New("user: password is too short, minimum password length is 8 characters")
	ErrPasswordTooLong   = errors.New("user: password is too long, maximum password length is 36 characters")
	// Session errors.
	ErrSessionExpired          = errors.New("session: session expired")
	ErrSessionUserIDEmpty      = errors.New("session: user ID is empty")
	ErrSessionRandomIDEmpty    = errors.New("session: random ID is empty")
	ErrSessionCreatedAtInvalid = errors.New("session: created at timestamp invalid")
	ErrSessionCreatedAtTooOld  = errors.New("session: created at timestamp is too old")
)
