package termretry

import (
	"time"

	retrygo "github.com/CHE3MZ/terminull/v2/termretry/internal/retry"
)

// Options tunes Do. Zero value retries 3 times with 1s backoff.
type Options struct {
	// Attempts caps tries including the first call. Default 3.
	Attempts uint
	// Delay is the base backoff between tries. Default time.Second.
	Delay time.Duration
	// RetryIf, when set, vetoes retries for matching errors.
	RetryIf func(error) bool
}

// Do runs fn until it succeeds or attempts run out, backing off between
// tries. It returns nil on success, otherwise the last error unwrapped.
func Do(fn func() error, opts Options) error {
	attempts := opts.Attempts
	if attempts == 0 {
		attempts = 3
	}
	delay := opts.Delay
	if delay <= 0 {
		delay = time.Second
	}

	args := []retrygo.Option{
		retrygo.Attempts(attempts),
		retrygo.Delay(delay),
		retrygo.DelayType(retrygo.BackOffDelay),
		retrygo.LastErrorOnly(true),
	}
	if opts.RetryIf != nil {
		args = append(args, retrygo.RetryIf(opts.RetryIf))
	}
	return retrygo.Do(fn, args...)
}
