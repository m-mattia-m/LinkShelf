package controller

import (
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

var installErrorLoggingOnce sync.Once

// installErrorLogging logs every API error that wraps an underlying Go error.
// huma.NewError is the choke point for all huma.ErrorXXX helpers, so wrapping
// it covers every controller. Errors without a cause and huma's own request
// validation errors are expected outcomes and are not logged.
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
