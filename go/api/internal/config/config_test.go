package config

import (
	"os"
	"testing"
)

// setEnv sets the given env vars for the duration of the test and registers a
// cleanup that restores the previous values.
func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		old, ok := os.LookupEnv(k)
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(k, old)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

func baseEnv() map[string]string {
	return map[string]string{
		"LIMEN_SECRET":             "0123456789abcdef0123456789abcdef", // 32 bytes
		"EARTHED_COOKIE_SECURE":     "true",
		"EARTHED_COOKIE_SAMESITE":   "none",
		"EARTHED_DATABASE_URL":      "postgres://earthed:earthed@localhost:5432/earthed?sslmode=disable",
		"EARTHED_HTTP_ADDR":         ":8080",
		"EARTHED_BASE_URL":          "http://localhost:8080",
		"EARTHED_WEB_URL":           "http://localhost:3000",
		"EARTHED_TRUSTED_ORIGINS":   "",
		"EARTHED_LOG_FORMAT":        "pretty",
		"EARTHED_POLLING_FREQUENCY": "60s",
		"EARTHED_BATCH_SIZE":        "100",
		"EARTHED_WORKER_POOL_SIZE":  "5",
	}
}

// When TrustedOrigins is unset, the web URL must NOT be merged: the allowlist
// stays nil so the CORS middleware stays permissive (any origin allowed).
func TestLoadTrustedOriginsUnstaysPermissive(t *testing.T) {
	setEnv(t, baseEnv())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TrustedOrigins != nil {
		t.Fatalf("expected nil TrustedOrigins when unset, got %v", cfg.TrustedOrigins)
	}
}

// When TrustedOrigins is set without the web URL, Load must add the web URL.
func TestLoadTrustedOriginsMergesWebURL(t *testing.T) {
	env := baseEnv()
	env["EARTHED_TRUSTED_ORIGINS"] = "https://app.example.com"
	env["EARTHED_WEB_URL"] = "https://example.com"
	setEnv(t, env)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := "https://example.com"
	found := false
	for _, o := range cfg.TrustedOrigins {
		if o == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("web URL %q not merged into TrustedOrigins %v", want, cfg.TrustedOrigins)
	}
}

// A web URL with a trailing slash must be normalized; an existing entry with a
// trailing slash must not cause a duplicate.
func TestLoadTrustedOriginsNormalizesTrailingSlash(t *testing.T) {
	env := baseEnv()
	env["EARTHED_WEB_URL"] = "https://example.com/"
	env["EARTHED_TRUSTED_ORIGINS"] = "https://example.com/"
	setEnv(t, env)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	count := 0
	for _, o := range cfg.TrustedOrigins {
		if o == "https://example.com" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one normalized %q entry, got %d in %v", "https://example.com", count, cfg.TrustedOrigins)
	}
}
