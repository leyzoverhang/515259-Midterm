package user

import "errors"

var (
	ErrUserNotFound error = errors.New("user not found")
	ErrInvalidInput error = errors.New("invalid input")
)
