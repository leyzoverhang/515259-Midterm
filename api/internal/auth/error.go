package auth

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrInvalidState   = errors.New("invalid or expired state")
	ErrInvalidTicket  = errors.New("invalid or expired ticket")
	ErrExchangeFailed = errors.New("exchange code failed")
)
