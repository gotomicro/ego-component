package consumerserver

import "errors"

var ErrRecoverableError error = errors.New("recoverable error is retryable")
