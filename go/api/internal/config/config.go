// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	BaseURL        string
	WebURL         string
	LimenSecret    string
	LogFormat      string
	PollingFreq    time.Duration
	BatchSize      int
	WorkerPool     int
	HTTPTimeout    time.Duration
	HTTPMaxBody    int64
	CleanupFreq    time.Duration
	EntryMaxAge    int
	DisableSched   bool
	CookieSecure   bool
	CookieSameSite string
	TrustedOrigins []string
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:       env("EARTHED_HTTP_ADDR", ":8080"),
		DatabaseURL:    env("EARTHED_DATABASE_URL", "postgres://earthed:earthed@localhost:5432/earthed?sslmode=disable"),
		BaseURL:        env("EARTHED_BASE_URL", "http://localhost:8080"),
		WebURL:         env("EARTHED_WEB_URL", "http://localhost:3000"),
		LimenSecret:    env("LIMEN_SECRET", ""),
		LogFormat:      env("EARTHED_LOG_FORMAT", "pretty"),
		PollingFreq:    envDuration("EARTHED_POLLING_FREQUENCY", 60*time.Second),
		BatchSize:      envInt("EARTHED_BATCH_SIZE", 100),
		WorkerPool:     envInt("EARTHED_WORKER_POOL_SIZE", 5),
		HTTPTimeout:    envDuration("EARTHED_HTTP_CLIENT_TIMEOUT", 20*time.Second),
		HTTPMaxBody:    int64(envInt("EARTHED_HTTP_CLIENT_MAX_BODY", 15*1024*1024)),
		CleanupFreq:    envDuration("EARTHED_CLEANUP_FREQUENCY", 24*time.Hour),
		EntryMaxAge:    envInt("EARTHED_ENTRY_MAX_AGE_DAYS", 60),
		DisableSched:   env("EARTHED_DISABLE_SCHEDULER", "") != "",
		CookieSecure:   envBool("EARTHED_COOKIE_SECURE", false),
		CookieSameSite: env("EARTHED_COOKIE_SAMESITE", "lax"),
		TrustedOrigins: envList("EARTHED_TRUSTED_ORIGINS"),
	}

	if cfg.LimenSecret == "" {
		return nil, fmt.Errorf("LIMEN_SECRET must be set (32 bytes)")
	}
	if len(cfg.LimenSecret) != 32 {
		return nil, fmt.Errorf("LIMEN_SECRET must be exactly 32 bytes, got %d", len(cfg.LimenSecret))
	}

	switch cfg.CookieSameSite {
	case "lax", "none", "strict":
	default:
		return nil, fmt.Errorf("EARTHED_COOKIE_SAMESITE must be one of lax, none, strict; got %q", cfg.CookieSameSite)
	}
	// SameSite=None requires Secure, otherwise browsers reject the cookie.
	if cfg.CookieSameSite == "none" && !cfg.CookieSecure {
		return nil, fmt.Errorf("EARTHED_COOKIE_SAMESITE=none requires EARTHED_COOKIE_SECURE=true (browsers reject SameSite=None without Secure)")
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

func envBool(key string, fallback bool) bool {
	switch os.Getenv(key) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

// envList reads a comma-separated env var into a trimmed slice. Returns nil
// when unset or empty so callers can distinguish "not configured" from
// "configured as one empty entry".
func envList(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
