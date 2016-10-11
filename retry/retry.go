package retry

import (
	"time"

	"github.com/cenk/backoff"
	"github.com/uber-go/zap"
)

const defaultMaxElapsedTime = 10 * time.Second
const defaultMaxInterval = 500 * time.Millisecond

func Retrying(operation func() error, logger zap.Logger) error {
	notify := func(err error, duration time.Duration) {
		logger.Warn(
			"Encountered error during run",
			zap.Int64("duration", duration.Nanoseconds()/int64(time.Millisecond)),
			zap.Error(err),
		)
	}

	customBackoff := backoff.NewExponentialBackOff()
	customBackoff.MaxElapsedTime = defaultMaxElapsedTime
	customBackoff.MaxInterval = defaultMaxInterval

	return backoff.RetryNotify(operation, customBackoff, notify)
}
