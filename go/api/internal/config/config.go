// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	BaseURL      string
	WebURL       string
	LimenSecret  string
	LogFormat    string
	PollingFreq  time.Duration
	BatchSize    int
	WorkerPool   int
	HTTPTimeout  time.Duration
	HTTPMaxBody  int64
	CleanupFreq  time.Duration
	EntryMaxAge  int
	DisableSched bool
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:     env("EARTHED_HTTP_ADDR", ":8080"),
		DatabaseURL:  env("EARTHED_DATABASE_URL", "postgres://earthed:earthed@localhost:5432/earthed?sslmode=disable"),
		BaseURL:      env("EARTHED_BASE_URL", "http://localhost:8080"),
		WebURL:       env("EARTHED_WEB_URL", "http://localhost:3000"),
		LimenSecret:  env("LIMEN_SECRET", ""),
		LogFormat:    env("EARTHED_LOG_FORMAT", "pretty"),
		PollingFreq:  envDuration("EARTHED_POLLING_FREQUENCY", 60*time.Second),
		BatchSize:    envInt("EARTHED_BATCH_SIZE", 100),
		WorkerPool:   envInt("EARTHED_WORKER_POOL_SIZE", 5),
		HTTPTimeout:  envDuration("EARTHED_HTTP_CLIENT_TIMEOUT", 20*time.Second),
		HTTPMaxBody:  int64(envInt("EARTHED_HTTP_CLIENT_MAX_BODY", 15*1024*1024)),
		CleanupFreq:  envDuration("EARTHED_CLEANUP_FREQUENCY", 24*time.Hour),
		EntryMaxAge:  envInt("EARTHED_ENTRY_MAX_AGE_DAYS", 60),
		DisableSched: env("EARTHED_DISABLE_SCHEDULER", "") != "",
	}

	if cfg.LimenSecret == "" {
		return nil, fmt.Errorf("LIMEN_SECRET must be set (32 bytes)")
	}
	if len(cfg.LimenSecret) != 32 {
		return nil, fmt.Errorf("LIMEN_SECRET must be exactly 32 bytes, got %d", len(cfg.LimenSecret))
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}
