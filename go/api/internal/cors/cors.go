// Package cors provides a permissive CORS middleware for the Earthed HTTP API.
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

// Middleware enables permissive CORS for all origins.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		if origin := r.Header.Get("Origin"); origin != "" {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Add("Vary", "Origin")
		} else {
			h.Set("Access-Control-Allow-Origin", "*")
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
