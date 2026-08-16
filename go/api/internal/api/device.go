package api

import (
	"context"
	"crypto/rand"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/fuegoio/planetary/go/api/internal/auth"
	"github.com/fuegoio/planetary/go/api/internal/ratelimit"
)

// Device-flow tuning constants.
const (
	deviceCodeTTL      = 5 * time.Minute                   // grant lifetime
	devicePollInterval = 5                                 // seconds, returned to the CLI
	deviceTokenTTL     = 14 * 24 * time.Hour               // 14 days, stored on the issued api token
	userCodeAlphabet   = "ABCDEFGHJKMNPQRSTUVWXYZ23456789" // no ambiguous chars (0/O, 1/I/L)
	userCodeLen        = 8
)

// Rate limits (per IP, in-memory, single-instance).
const (
	rateDeviceCode  = 10.0 / 60.0 // 10 per minute
	burstDeviceCode = 10
	rateConfirm     = 20.0 / 60.0 // 20 per minute
	burstConfirm    = 20
)

// PublicDevicePaths is the set of device-flow paths mounted outside the
// auth middleware (mirrors /api/v1/health). main.go reads this.
var PublicDevicePaths = []string{
	"/api/auth/device/code",
	"/api/auth/device/token",
}

// registerDeviceRoutes registers the RFC 8628 device-authorization grant
// endpoints under /api/auth/device/*. The public issue/poll endpoints are
// mounted outside the auth middleware in main.go (like /api/v1/health);
// the confirm/status endpoints sit behind the session-cookie middleware.
func (a *API) registerDeviceRoutes() {
	limiter := ratelimit.New()

	// --- POST /api/auth/device/code (public) ---
	huma.Register(a.huma, huma.Operation{
		OperationID: "device-code",
		Method:      http.MethodPost,
		Path:        "/api/auth/device/code",
		Summary:     "Begin device-flow login (issue a device code)",
		Tags:        []string{"device"},
	}, func(ctx context.Context, input *DeviceCodeInput) (*DeviceCodeOutput, error) {
		ip := clientIP(input)
		if !limiter.Allow(ip, rateDeviceCode, burstDeviceCode) {
			return nil, huma.Error429TooManyRequests("rate limit exceeded")
		}

		rawDeviceCode := generateToken() // reuse pla_<32 bytes hex> generator for entropy
		dc, err := a.store.CreateDeviceCode(ctx, auth.HashToken(rawDeviceCode), generateUserCode(), devicePollInterval, time.Now().Add(deviceCodeTTL))
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		webURL := a.webURL()
		return &DeviceCodeOutput{Body: struct {
			DeviceCode              string `json:"device_code"`
			UserCode                string `json:"user_code"`
			VerificationURI         string `json:"verification_uri"`
			VerificationURIComplete string `json:"verification_uri_complete"`
			ExpiresIn               int    `json:"expires_in"`
			Interval                int    `json:"interval"`
		}{
			DeviceCode:              rawDeviceCode,
			UserCode:                dc.UserCode,
			VerificationURI:         webURL + "/device",
			VerificationURIComplete: webURL + "/device?user_code=" + dc.UserCode,
			ExpiresIn:               int(deviceCodeTTL.Seconds()),
			Interval:                devicePollInterval,
		}}, nil
	})

	// --- POST /api/auth/device/token (public, polled by the CLI) ---
	huma.Register(a.huma, huma.Operation{
		OperationID: "device-token",
		Method:      http.MethodPost,
		Path:        "/api/auth/device/token",
		Summary:     "Poll for device-flow login result",
		Tags:        []string{"device"},
	}, func(ctx context.Context, input *DeviceTokenInput) (*DeviceTokenOutput, error) {
		ip := clientIP(input)
		if !limiter.Allow(ip, rateDeviceCode, burstDeviceCode) {
			return nil, huma.Error429TooManyRequests("rate limit exceeded")
		}

		dc, err := a.store.GetDeviceCodeByHash(ctx, auth.HashToken(input.Body.DeviceCode))
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if dc == nil || dc.Status == "expired" {
			return nil, deviceError("expired_token")
		}

		// slow_down: client polled faster than the advertised interval.
		if dc.LastPolledAt != nil && time.Since(*dc.LastPolledAt) < time.Duration(dc.IntervalSecs)*time.Second {
			return nil, deviceError("slow_down")
		}
		if err := a.store.TouchDeviceCodePoll(ctx, dc.DeviceCode); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		switch dc.Status {
		case "pending":
			return nil, deviceError("authorization_pending")
		case "denied":
			return nil, deviceError("access_denied")
		case "authorized":
			if dc.TokenPlaintext == nil || *dc.TokenPlaintext == "" {
				return nil, deviceError("expired_token")
			}
			token := *dc.TokenPlaintext
			// Single-use: delete the grant so the code can't be replayed.
			_ = a.store.ConsumeDeviceCode(ctx, dc.DeviceCode)
			return &DeviceTokenOutput{Body: struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
				ExpiresIn   int    `json:"expires_in"`
			}{
				AccessToken: token,
				TokenType:   "Bearer",
				ExpiresIn:   int(deviceTokenTTL.Seconds()),
			}}, nil
		}
		return nil, deviceError("authorization_pending")
	})

	// --- POST /api/auth/device/confirm (session-cookie auth) ---
	huma.Register(a.huma, huma.Operation{
		OperationID: "device-confirm",
		Method:      http.MethodPost,
		Path:        "/api/auth/device/confirm",
		Summary:     "Approve or deny a device-flow login (authenticated user)",
		Tags:        []string{"device"},
	}, func(ctx context.Context, input *DeviceConfirmInput) (*DeviceConfirmOutput, error) {
		ip := clientIP(input)
		if !limiter.Allow(ip, rateConfirm, burstConfirm) {
			return nil, huma.Error429TooManyRequests("rate limit exceeded")
		}

		userID := auth.UserIDFromCtx(ctx)

		dc, err := a.store.GetDeviceCodeByUserCode(ctx, strings.ToUpper(input.Body.UserCode))
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if dc == nil {
			return nil, huma.Error404NotFound("device code not found")
		}
		if dc.Status == "expired" {
			return nil, huma.Error400BadRequest("device code expired")
		}
		if dc.Status != "pending" {
			return nil, huma.Error409Conflict("device code already used")
		}

		if input.Body.Deny {
			if _, err := a.store.DenyDeviceCode(ctx, dc.UserCode); err != nil {
				return nil, huma.Error500InternalServerError(err.Error())
			}
			return &DeviceConfirmOutput{Body: DeviceConfirmBody{Authorized: false}}, nil
		}

		// Mint the API token now. The hash goes to api_tokens for ongoing
		// bearer auth; the plaintext is stashed on the device_codes row and
		// returned once to the polling CLI, then the row is deleted.
		rawToken := generateToken()
		hash := auth.HashToken(rawToken)
		t, err := a.store.CreateAPIToken(ctx, userID, "CLI (device flow)", hash, "device_flow", timePtr(time.Now().Add(deviceTokenTTL)))
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if _, err := a.store.AuthorizeDeviceCode(ctx, dc.UserCode, userID, t.ID, rawToken); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		return &DeviceConfirmOutput{Body: DeviceConfirmBody{Authorized: true}}, nil
	})

	// --- GET /api/auth/device/status (session-cookie auth) ---
	huma.Register(a.huma, huma.Operation{
		OperationID: "device-status",
		Method:      http.MethodGet,
		Path:        "/api/auth/device/status",
		Summary:     "Check device-flow status by user code (authenticated user)",
		Tags:        []string{"device"},
	}, func(ctx context.Context, input *DeviceStatusInput) (*DeviceStatusOutput, error) {
		ip := clientIP(input)
		if !limiter.Allow(ip, rateConfirm, burstConfirm) {
			return nil, huma.Error429TooManyRequests("rate limit exceeded")
		}

		dc, err := a.store.GetDeviceCodeByUserCode(ctx, strings.ToUpper(input.UserCode))
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		status := "not_found"
		if dc != nil {
			status = dc.Status
		}
		return &DeviceStatusOutput{Body: DeviceStatusBody{Status: status}}, nil
	})
}

// --- input / output types ---

type DeviceCodeInput struct{}

type DeviceCodeOutput struct {
	Body struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
}

type DeviceTokenInput struct {
	Body struct {
		DeviceCode string `json:"device_code" minLength:"1"`
	}
}

type DeviceTokenOutput struct {
	Body struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
}

type DeviceConfirmInput struct {
	Body struct {
		UserCode string `json:"user_code" minLength:"1" maxLength:"32"`
		Deny     bool   `json:"deny,omitempty"`
	}
}

type DeviceConfirmBody struct {
	Authorized bool `json:"authorized"`
}

type DeviceConfirmOutput struct {
	Body DeviceConfirmBody
}

type DeviceStatusInput struct {
	UserCode string `query:"user_code" maxLength:"32"`
}

type DeviceStatusBody struct {
	Status string `json:"status"`
}

type DeviceStatusOutput struct {
	Body DeviceStatusBody
}

// --- helpers ---

// webURL returns the configured web origin for building verification URIs.
// Defaults to localhost; set PLANETARY_WEB_URL for cloud deployments.
func (a *API) webURL() string {
	if a.cfg != nil && a.cfg.WebURL != "" {
		return strings.TrimRight(a.cfg.WebURL, "/")
	}
	return "http://localhost:3000"
}

// generateUserCode returns a human-readable code like "PLN-XXXX-XXXX".
func generateUserCode() string {
	b := make([]byte, userCodeLen)
	_, _ = rand.Read(b)
	out := make([]byte, 0, 4+userCodeLen+1)
	out = append(out, 'P', 'L', 'N', '-')
	for i, c := range b {
		out = append(out, userCodeAlphabet[int(c)%len(userCodeAlphabet)])
		if i == 3 {
			out = append(out, '-')
		}
	}
	return string(out)
}

// deviceError returns a huma error whose body is a JSON object of the form
// {"error":"<code>"} matching RFC 8628 polling error semantics.
func deviceError(code string) huma.StatusError {
	return huma.Error400BadRequest(`{"error":"` + code + `"}`)
}

// clientIP extracts the caller IP for rate limiting. huma input structs do
// not expose the request directly; we read X-Forwarded-For via the huma
// header accessor when present, otherwise fall back to "unknown". In a
// single-node deployment the rate limiter still bounds total load; behind
// a proxy, set X-Forwarded-For for per-client limits.
func clientIP(input any) string {
	type headerer interface {
		Headers() http.Header
	}
	if h, ok := input.(headerer); ok {
		if xff := h.Headers().Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i > 0 {
				return strings.TrimSpace(xff[:i])
			}
			return strings.TrimSpace(xff)
		}
	}
	return "unknown"
}

// timePtr returns a pointer to t.
func timePtr(t time.Time) *time.Time {
	return &t
}
