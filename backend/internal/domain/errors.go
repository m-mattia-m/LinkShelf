package domain

import (
	"errors"
	"strings"
)

var (
	// ErrForbidden: authenticated, but neither owner nor admin.
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound means the requested resource (or one it depends on) doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidRole means a caller tried to set a role other than "user" or "admin".
	ErrInvalidRole = errors.New("invalid role")
	// ErrInvalidInput means a request field failed a business-rule validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrConflict: clashes with an existing resource, e.g. a taken username.
	ErrConflict = errors.New("conflict")
	// ErrPasswordResetDisabled: authentication.passwordReset.enabled is false.
	ErrPasswordResetDisabled = errors.New("password reset is currently disabled")
	// ErrRegistrationDisabled: authentication.registrationEnabled is false.
	ErrRegistrationDisabled = errors.New("registration is currently disabled")
	// ErrEmailVerificationPending: login requires a verified email (or a set password).
	ErrEmailVerificationPending = errors.New("this account's email address has not been verified yet")
	// ErrTooManyRequests: the same email action was requested within resendCooldown.
	ErrTooManyRequests = errors.New("please wait a minute before trying again")
	// ErrLocalAuthDisabled: authentication.localAuthEnabled is false.
	ErrLocalAuthDisabled = errors.New("password login is disabled on this instance")
)

// normalizeEmail trims and lowercases an email for storage and comparison.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
