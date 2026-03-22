package util

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"
)

// Retry calls fn up to maxAttempts times using exponential backoff with ±25%
// jitter between attempts. It returns on the first success or immediately when
// the context is cancelled. maxAttempts ≤ 1 calls fn exactly once (no retry).
//
// Backoff schedule (baseDelay = 100 ms, before jitter):
//
//	attempt 1 → immediate
//	attempt 2 → ~100 ms
//	attempt 3 → ~200 ms
func Retry[T any](
	ctx context.Context,
	maxAttempts int,
	baseDelay time.Duration,
	fn func() (T, error),
) (T, error) {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var (
		zero    T
		lastErr error
	)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := retryJitter(baseDelay * time.Duration(1<<(attempt-1)))
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		lastErr = err
		slog.WarnContext(ctx, "retry attempt failed",
			"attempt", attempt+1,
			"max_attempts", maxAttempts,
			"error", err,
		)
	}

	return zero, lastErr
}

// retryJitter scales d by a random factor in [0.75, 1.25).
func retryJitter(d time.Duration) time.Duration {
	factor := 0.75 + rand.Float64()*0.5
	return time.Duration(float64(d) * factor)
}
