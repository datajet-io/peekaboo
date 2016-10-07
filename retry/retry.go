package retry

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
)

const defaultMaxElapsedTime = 10 * time.Second
const defaultMaxInterval = 500 * time.Millisecond

func Retrying(operation func() error) error {
	notify := func(err error, duration time.Duration) {
		fmt.Printf("Received error during run: %s\n", err)
	}

	customBackoff := backoff.NewExponentialBackOff()
	customBackoff.MaxElapsedTime = defaultMaxElapsedTime
	customBackoff.MaxInterval = defaultMaxInterval

	return backoff.RetryNotify(operation, customBackoff, notify)
}
