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
	// ErrInvalidInput means a request field failed a business-rule validation
	// (as opposed to the basic required/type checks huma already enforces).
	ErrInvalidInput = errors.New("invalid input")
	// ErrRegistrationDisabled means a caller tried to self-register while
	// authentication.registrationEnabled is false. Admin-created accounts
	// and OIDC auto-provisioning are unaffected by this toggle.
	ErrRegistrationDisabled = errors.New("registration is currently disabled")
	// ErrEmailVerificationPending means the account exists and the supplied
	// password (if any) checked out, but it can't log in yet because its
	// email hasn't been verified (or, for an admin-invited account, no
	// password has been set yet). A fresh link is sent whenever this is hit.
	ErrEmailVerificationPending = errors.New("this account's email address has not been verified yet")
)
