package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTransient = errors.New("transient error")

// callN returns a function that fails the first n calls then returns the value.
func callN[T any](fails int, value T) func() (T, error) {
	left := fails
	return func() (T, error) {
		if left > 0 {
			left--
			var zero T
			return zero, errTransient
		}
		return value, nil
	}
}

func TestRetry_SucceedsOnFirstAttempt(t *testing.T) {
	calls := 0
	result, err := Retry(context.Background(), 3, time.Millisecond, func() (string, error) {
		calls++
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 1, calls)
}

func TestRetry_SucceedsAfterOneFailure(t *testing.T) {
	fn := callN(1, "value")
	result, err := Retry(context.Background(), 3, time.Millisecond, fn)
	require.NoError(t, err)
	assert.Equal(t, "value", result)
}

func TestRetry_SucceedsAfterTwoFailures(t *testing.T) {
	fn := callN(2, 42)
	result, err := Retry(context.Background(), 3, time.Millisecond, fn)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestRetry_ExhaustsAllAttempts(t *testing.T) {
	calls := 0
	_, err := Retry(context.Background(), 3, time.Millisecond, func() (int, error) {
		calls++
		return 0, errTransient
	})
	assert.ErrorIs(t, err, errTransient)
	assert.Equal(t, 3, calls)
}

func TestRetry_MaxAttemptsZeroMeansOnce(t *testing.T) {
	calls := 0
	_, err := Retry(context.Background(), 0, time.Millisecond, func() (int, error) {
		calls++
		return 0, errTransient
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetry_MaxAttemptsOneMeansOnce(t *testing.T) {
	calls := 0
	_, err := Retry(context.Background(), 1, time.Millisecond, func() (int, error) {
		calls++
		return 0, errTransient
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetry_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	_, err := Retry(ctx, 5, time.Millisecond, func() (string, error) {
		calls++
		if calls >= 2 {
			cancel()
		}
		return "", errTransient
	})

	assert.ErrorIs(t, err, context.Canceled)
	assert.LessOrEqual(t, calls, 3)
}

func TestRetry_AlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before first attempt

	calls := 0
	_, err := Retry(ctx, 3, time.Millisecond, func() (string, error) {
		calls++
		return "", errTransient
	})

	// First attempt still runs; then the wait detects ctx.Done or ctx.Err check fires.
	assert.Error(t, err)
}

func TestRetry_ContextCancelledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	// Use a long backoff so the ctx.Done() select case fires before the timer.
	_, err := Retry(ctx, 3, 500*time.Millisecond, func() (string, error) {
		calls++
		cancel() // cancel immediately after first fn() call
		return "", errTransient
	})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls, "should stop after first attempt once context is cancelled during backoff")
}

func TestRetry_ContextCancelledDuringBackoffWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	_, err := Retry(ctx, 3, 200*time.Millisecond, func() (string, error) {
		calls++
		if calls == 1 {
			// Schedule cancellation to fire during the backoff sleep, not during fn()
			go func() {
				time.Sleep(30 * time.Millisecond)
				cancel()
			}()
		}
		return "", errTransient
	})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls, "should not retry after context is cancelled during the backoff wait")
}
