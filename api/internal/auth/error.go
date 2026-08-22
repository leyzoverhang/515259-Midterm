package auth

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrInvalidState   = errors.New("invalid or expired state")
	ErrExchangeFailed = errors.New("exchange code failed")
)
