package domain

import (
	"errors"
	"strings"
)

var (
	// ErrForbidden means the caller is authenticated but not allowed to act on
	// the requested resource (not its owner, and not an admin).
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound means the requested resource (or one it depends on) doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidRole means a caller tried to set a role other than "user" or "admin".
	ErrInvalidRole = errors.New("invalid role")
	// ErrInvalidInput means a request field failed a business-rule validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrConflict means the request is valid but clashes with something that
	// already exists, such as a taken username or shelf path.
	ErrConflict = errors.New("conflict")
	// ErrPasswordResetDisabled means a caller used the forgot-password flow
	// while authentication.passwordReset.enabled is false.
	ErrPasswordResetDisabled = errors.New("password reset is currently disabled")
	// ErrRegistrationDisabled means a caller tried to self-register while
	// authentication.registrationEnabled is false.
	ErrRegistrationDisabled = errors.New("registration is currently disabled")
	// ErrEmailVerificationPending means the account can't log in until its
	// email is verified (or, if admin-invited, a password is set).
	ErrEmailVerificationPending = errors.New("this account's email address has not been verified yet")
	// ErrTooManyRequests means the same email action was requested again
	// within resendCooldown and was refused instead of silently dropped,
	// because the caller must know no new link was sent.
	ErrTooManyRequests = errors.New("please wait a minute before trying again")
	// ErrLocalAuthDisabled means a caller used password login, sign-up or
	// password reset while authentication.localAuthEnabled is false.
	ErrLocalAuthDisabled = errors.New("password login is disabled on this instance")
)

// normalizeEmail is how every email is stored and compared: trimmed and
// lowercased, so two spellings of one address can never be two accounts.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
