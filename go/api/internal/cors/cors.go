// Package cors provides a CORS middleware for the Sunred HTTP API.
package cors

import (
	"net/http"
	"strings"
)

var allowedHeaders = []string{
	"Accept",
	"Authorization",
	"Content-Type",
	"X-Requested-With",
}

var allowedMethods = []string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"OPTIONS",
}

// Middleware enables CORS. When origins is empty it reflects any Origin back
// with Allow-Credentials (permissive, for local development). When origins
// is non-empty it only echoes a request Origin that matches the allowlist,
// so cross-domain credentialed requests are bounded to trusted frontends.
func Middleware(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = true
	}
	permissive := len(allowed) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			origin := r.Header.Get("Origin")

			switch {
			case origin != "" && (permissive || allowed[origin]):
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Add("Vary", "Origin")
			case origin == "":
				h.Set("Access-Control-Allow-Origin", "*")
			default:
				// Untrusted origin: send no allow-origin header. Vary by Origin
				// so caches don't poison responses for other origins.
				h.Add("Vary", "Origin")
			}
			h.Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			h.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			h.Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
