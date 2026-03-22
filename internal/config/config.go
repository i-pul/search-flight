package config

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr string `envconfig:"ADDR" default:":8080"`

	// Maximum time in milliseconds to wait for all provider responses.
	ProviderTimeoutMs int `envconfig:"PROVIDER_TIMEOUT_MS" default:"2000"`

	// Retry configuration for failed provider calls.
	RetryMaxAttempts int `envconfig:"RETRY_MAX_ATTEMPTS" default:"3"`
	RetryBaseDelayMs int `envconfig:"RETRY_BASE_DELAY_MS" default:"100"`

	// Best-value scoring weights (must be positive; automatically normalised to sum to 1).
	BestValueWeightPrice    float64 `envconfig:"BEST_VALUE_WEIGHT_PRICE" default:"0.50"`
	BestValueWeightDuration float64 `envconfig:"BEST_VALUE_WEIGHT_DURATION" default:"0.30"`
	BestValueWeightStops    float64 `envconfig:"BEST_VALUE_WEIGHT_STOPS" default:"0.20"`
}

func Load() Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	return cfg
}
