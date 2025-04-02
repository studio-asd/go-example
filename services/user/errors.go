package user

import "errors"

var (
	ErrUserNotFound        = errors.New("user: not found")
	ErrUserSessionNotFound = errors.New("user: session not found")
	ErrSessionExpired      = errors.New("user: session expired")
	ErrInvalidPassword     = errors.New("user: invalid password")
)
