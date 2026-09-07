package domain

import "errors"

var (
	// ErrForbidden means the caller is authenticated but not allowed to act on
	// the requested resource (not its owner, and not an admin).
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound means the requested resource (or one it depends on) doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidRole means a caller tried to set a role other than "user" or "admin".
	ErrInvalidRole = errors.New("invalid role")
)
