package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_UsesDefaults(t *testing.T) {
	// Ensure no env vars interfere.
	os.Unsetenv("ADDR")
	os.Unsetenv("PROVIDER_TIMEOUT_MS")
	os.Unsetenv("RETRY_MAX_ATTEMPTS")
	os.Unsetenv("RETRY_BASE_DELAY_MS")
	os.Unsetenv("BEST_VALUE_WEIGHT_PRICE")
	os.Unsetenv("BEST_VALUE_WEIGHT_DURATION")
	os.Unsetenv("BEST_VALUE_WEIGHT_STOPS")

	cfg := Load()

	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, 2000, cfg.ProviderTimeoutMs)
	assert.Equal(t, 3, cfg.RetryMaxAttempts)
	assert.Equal(t, 100, cfg.RetryBaseDelayMs)
	assert.Equal(t, 0.50, cfg.BestValueWeightPrice)
	assert.Equal(t, 0.30, cfg.BestValueWeightDuration)
	assert.Equal(t, 0.20, cfg.BestValueWeightStops)
}

func TestLoad_ReadsEnvironmentVariables(t *testing.T) {
	os.Setenv("ADDR", ":9090")
	os.Setenv("PROVIDER_TIMEOUT_MS", "3000")
	os.Setenv("RETRY_MAX_ATTEMPTS", "5")
	defer func() {
		os.Unsetenv("ADDR")
		os.Unsetenv("PROVIDER_TIMEOUT_MS")
		os.Unsetenv("RETRY_MAX_ATTEMPTS")
	}()

	cfg := Load()

	assert.Equal(t, ":9090", cfg.Addr)
	assert.Equal(t, 3000, cfg.ProviderTimeoutMs)
	assert.Equal(t, 5, cfg.RetryMaxAttempts)
}
