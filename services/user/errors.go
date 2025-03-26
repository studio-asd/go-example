package user

import "errors"

var (
	ErrUserSessionNotFound = errors.New("user: session not found")
	ErrSessionExpired      = errors.New("user: session expired")
)
