package controller

import (
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

var installErrorLoggingOnce sync.Once

// installErrorLogging makes every API error response that carries an
// underlying Go error get logged server-side, exactly once, no matter which
// controller produced it. Without this, a controller that does
// `huma.Error400BadRequest("failed to X", err)` and nothing else leaves no
// trace anywhere but the HTTP response - if the client doesn't surface it
// (or discards it, as a best-effort call does), the failure is invisible.
//
// huma.NewError is the single choke point every huma.ErrorXXX helper calls
// through, so wrapping it here covers the whole API, present and future,
// without touching each of the many call sites individually.
//
// Only errors that carry an attached cause (errs) are logged - a plain
// "not found" / "forbidden" / validation message with no wrapped error is an
// expected outcome, not a failure, and logging every one of those at error
// level would drown out the failures that actually matter. huma's own
// request-body validation errors (msg == "validation failed") are skipped
// for the same reason: a malformed request is a client mistake, not a
// backend failure.
//
// Callers must never pass credentials, passwords, or tokens as the wrapped
// error - only errors that already exist in this codebase reach here, and
// none of them do that (see mailer/email_verification for the case this was
// added for).
func installErrorLogging() {
	installErrorLoggingOnce.Do(func() {
		next := huma.NewError
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			if msg != "validation failed" {
				for _, err := range errs {
					if err != nil {
						zap.L().Error(msg, zap.Int("status", status), zap.Error(err))
					}
				}
			}
			return next(status, msg, errs...)
		}
	})
}
