package controller

import (
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

var installErrorLoggingOnce sync.Once

// installErrorLogging logs API errors that wrap a Go error by wrapping huma.NewError.
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
