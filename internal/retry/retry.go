// Package retry provides a simple exponential-backoff retry helper for
// transient IBM Verify API errors (429, 5xx, network blips).
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
)

const (
	maxAttempts = 3
	baseDelay   = 1 * time.Second
	maxDelay    = 10 * time.Second
)

// isRetryable reports whether the error warrants a retry.
// We only retry idempotent failures: rate-limits and server-side 5xx.
// Auth (401/403), not-found (404), and validation errors are permanent.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRateLimit() || apiErr.IsServer()
	}
	// network-level errors (connection refused, timeout) are also retryable
	return true
}

// Do calls fn up to maxAttempts times. On a retryable failure it waits with
// exponential backoff (1s, 2s … capped at 10s) before the next attempt.
// The context is checked before each sleep so cancellation is respected.
// Non-retryable errors (auth, not-found, validation) are returned immediately.
func Do(ctx context.Context, fn func() error) error {
	return doWithDelay(ctx, fn, baseDelay)
}

// doWithDelay is the internal implementation that accepts an initial delay so
// tests can pass 0 to avoid real sleeps.
func doWithDelay(ctx context.Context, fn func() error, initial time.Duration) error {
	delay := initial
	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if attempt >= maxAttempts || !isRetryable(err) {
			return err
		}
		if delay > maxDelay {
			delay = maxDelay
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
}
