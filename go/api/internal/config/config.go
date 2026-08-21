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
	CookieDomain   string
	TrustedOrigins []string
	// RelayURL is the base URL of the Earthed relay. When set, the API
	// announces new AT Proto users to the relay and queries it for global counts.
	// Leave empty to disable relay integration.
	RelayURL string
	// DefaultPDS is the PDS URL used for new account signup. When set, the web
	// login page shows a "Create account" link that starts the OAuth flow
	// against this PDS. When unset, only existing-handle login is offered.
	DefaultPDS string
	// OAuthClientID is the full URL where Earthed serves its client metadata
	// document (the PDS fetches this during the OAuth flow). Defaults to
	// "<BaseURL>/client-metadata.json".
	OAuthClientID string
	// OAuthCallbackURL is the public OAuth redirect URL the PDS sends the
	// authorization code back to. Defaults to "<BaseURL>/auth/oauth/callback".
	OAuthCallbackURL string
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:       env("EARTHED_HTTP_ADDR", ":8080"),
		DatabaseURL:    env("EARTHED_DATABASE_URL", "postgres://earthed:earthed@localhost:5432/earthed?sslmode=disable"),
		BaseURL:        env("EARTHED_BASE_URL", "http://localhost:8080"),
		WebURL:         env("EARTHED_WEB_URL", "http://localhost:3000"),
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
		CookieDomain:   strings.TrimSpace(env("EARTHED_COOKIE_DOMAIN", "")),
		TrustedOrigins: envList("EARTHED_TRUSTED_ORIGINS"),
		RelayURL:       env("EARTHED_RELAY_URL", ""),
		DefaultPDS:    strings.TrimRight(strings.TrimSpace(env("EARTHED_DEFAULT_PDS", "")), "/"),
	}

	// OAuth client_id and callback URL default from BaseURL. The client_id is a
	// URL pointing at the client-metadata.json document this server serves.
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.OAuthClientID = env("EARTHED_OAUTH_CLIENT_ID", base+"/client-metadata.json")
	cfg.OAuthCallbackURL = env("EARTHED_OAUTH_REDIRECT_URL", base+"/auth/oauth/callback")

	switch cfg.CookieSameSite {
	case "lax", "none", "strict":
	default:
		return nil, fmt.Errorf("EARTHED_COOKIE_SAMESITE must be one of lax, none, strict; got %q", cfg.CookieSameSite)
	}
	// SameSite=None requires Secure, otherwise browsers reject the cookie.
	if cfg.CookieSameSite == "none" && !cfg.CookieSecure {
		return nil, fmt.Errorf("EARTHED_COOKIE_SAMESITE=none requires EARTHED_COOKIE_SECURE=true (browsers reject SameSite=None without Secure)")
	}

	// CookieDomain must be a bare host (e.g. "earthed.app"), not a URL.
	// When set, the session cookie is shared across that registrable domain and
	// its subdomains so the web app and API (on a subdomain) share the session.
	if cfg.CookieDomain != "" {
		if strings.Contains(cfg.CookieDomain, "://") || strings.ContainsAny(cfg.CookieDomain, "/:") {
			return nil, fmt.Errorf("EARTHED_COOKIE_DOMAIN must be a bare host (e.g. \"earthed.app\"), not a URL; got %q", cfg.CookieDomain)
		}
	}

	// When an explicit trusted-origins allowlist is set, always include the
	// configured web frontend so its credentialed requests are accepted, and
	// normalize entries (trim trailing slashes) so they match browser Origin
	// headers, which never carry a trailing slash.
	// When the allowlist is unset (nil) the CORS middleware is permissive and
	// allows any origin, so the web URL is already covered.
	if cfg.TrustedOrigins != nil {
		webOrigin := strings.TrimRight(strings.TrimSpace(cfg.WebURL), "/")
		seen := make(map[string]bool, len(cfg.TrustedOrigins)+1)
		normalized := make([]string, 0, len(cfg.TrustedOrigins)+1)
		add := func(o string) {
			if !seen[o] {
				seen[o] = true
				normalized = append(normalized, o)
			}
		}
		for _, o := range cfg.TrustedOrigins {
			add(strings.TrimRight(strings.TrimSpace(o), "/"))
		}
		add(webOrigin)
		cfg.TrustedOrigins = normalized
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
