package retry

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
)

func Retrying(operation func() error) error {
	notify := func(err error, duration time.Duration) {
		fmt.Printf("Received error during run: %s\n", err)
	}

	customBackoff := backoff.NewExponentialBackOff()
	customBackoff.MaxElapsedTime = 10 * time.Second
	customBackoff.MaxInterval = 500 * time.Millisecond

	return backoff.RetryNotify(operation, customBackoff, notify)
}
