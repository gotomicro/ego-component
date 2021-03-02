package eredis

import (
	"time"
)

// RetryStrategy allows to customise the Lock retry strategy.
type RetryStrategy interface {
	// NextBackoff returns the next backoff duration.
	NextBackoff() time.Duration
}

// --------------------------------LinearBackoff Retry-----------------------------------
type linearBackoff time.Duration

// LinearBackoffRetry allows retries regularly with customized intervals
func LinearBackoffRetry(backoff time.Duration) RetryStrategy {
	return linearBackoff(backoff)
}

func (r linearBackoff) NextBackoff() time.Duration {
	return time.Duration(r)
}

// --------------------------------No Retry-----------------------------------
// NoRetry acquire the Lock only once.
func NoRetry() RetryStrategy {
	return linearBackoff(0)
}

// --------------------------------Limit Retry-----------------------------------
type limitedRetry struct {
	s RetryStrategy

	cnt, max int
}

// LimitRetry limits the number of retries to max attempts.
func LimitRetry(s RetryStrategy, max int) RetryStrategy {
	return &limitedRetry{s: s, max: max}
}

func (r *limitedRetry) NextBackoff() time.Duration {
	if r.cnt >= r.max {
		return 0
	}
	r.cnt++
	return r.s.NextBackoff()
}

// --------------------------------ExponentialBackoff Retry-----------------------------------
type exponentialBackoff struct {
	cnt      uint
	min, max time.Duration
}

// ExponentialBackoffRetry strategy is an optimization strategy with a retry time of 2**n milliseconds (n means number of times).
// You can set a minimum and maximum value, the recommended minimum value is not less than 16ms.
func ExponentialBackoffRetry(min, max time.Duration) RetryStrategy {
	return &exponentialBackoff{min: min, max: max}
}

func (r *exponentialBackoff) NextBackoff() time.Duration {
	r.cnt++
	ms := 2 << 25
	if r.cnt < 25 {
		ms = 2 << r.cnt
	}

	if d := time.Duration(ms) * time.Millisecond; d < r.min {
		return r.min
	} else if r.max != 0 && d > r.max {
		return r.max
	} else {
		return d
	}
}
