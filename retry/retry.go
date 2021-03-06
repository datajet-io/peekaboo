package retry

import (
	"time"

	"github.com/cenk/backoff"
	"go.uber.org/zap"
)

const defaultMaxInterval = 500 * time.Millisecond

// Retrying is a wrapper for cenk/backoff to handle exponential backoffs
func Retrying(operation func() error, retryTimeoutSeconds int, logger zap.Logger) error {
	notify := func(err error, duration time.Duration) {
		logger.Warn(
			"Encountered error during run",
			zap.Int64("duration", duration.Nanoseconds()/int64(time.Millisecond)),
			zap.Error(err),
		)
	}

	customBackoff := backoff.NewExponentialBackOff()
	customBackoff.MaxElapsedTime = time.Duration(retryTimeoutSeconds) * time.Second
	customBackoff.MaxInterval = defaultMaxInterval

	return backoff.RetryNotify(operation, customBackoff, notify)
}
