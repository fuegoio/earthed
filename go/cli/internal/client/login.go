package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// deviceCodeResponse is the body returned by POST /api/auth/device/code.
type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// deviceTokenResponse is the success body returned by POST /api/auth/device/token.
type deviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// deviceTokenErrorResponse is the RFC 8628 polling error body.
type deviceTokenErrorResponse struct {
	Error string `json:"error"`
}

// LoginResult is returned by a successful device-flow login.
type LoginResult struct {
	Token     string
	ExpiresIn int
}

// Login runs the device-flow login against the given API base URL. It prints
// the user code and verification URL, opens the browser if possible, and polls
// until the user approves, denies, or the grant expires. On success it returns
// the token to persist.
//
// openBrowser controls whether the verification URL is opened automatically;
// set it false for --no-browser or headless sessions. The URL and code are
// always printed regardless.
func Login(ctx context.Context, baseURL string, openBrowser bool, out io.Writer) (*LoginResult, error) {
	// 1. Request a device code.
	codeResp, err := requestDeviceCode(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	// 2. Show the code + URL, and open the browser if asked.
	fmt.Fprintf(out, "\nYour confirmation code:  %s\n\n", codeResp.UserCode)
	fmt.Fprintf(out, "Open this URL in a browser to approve:\n  %s\n\n", codeResp.VerificationURIComplete)
	if openBrowser {
		if err := openURL(codeResp.VerificationURIComplete); err != nil {
			fmt.Fprintf(out, "(could not open browser automatically: %v)\n", err)
		}
	}
	fmt.Fprintf(out, "Waiting for approval...")

	// 3. Poll until resolved.
	interval := time.Duration(codeResp.Interval) * time.Second
	if interval < 1*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(codeResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device code expired")
		}

		tokenResp, pollErr, wait := pollDeviceToken(ctx, baseURL, codeResp.DeviceCode)
		if pollErr != nil {
			switch pollErr.Error {
			case "authorization_pending":
				fmt.Fprint(out, ".")
				continue
			case "slow_down":
				interval *= 2
				fmt.Fprint(out, " (slowing down)")
				time.Sleep(interval)
				continue
			case "expired_token":
				return nil, fmt.Errorf("device code expired")
			case "access_denied":
				return nil, fmt.Errorf("login denied by user")
			default:
				return nil, fmt.Errorf("login failed: %s", pollErr.Error)
			}
		}
		if tokenResp != nil {
			fmt.Fprintf(out, "\nLogin successful.\n")
			return &LoginResult{
				Token:     tokenResp.AccessToken,
				ExpiresIn: tokenResp.ExpiresIn,
			}, nil
		}
		// wait hint from server (rare); fall back to interval.
		if wait > 0 {
			interval = time.Duration(wait) * time.Second
		}
		time.Sleep(interval)
	}
}

func requestDeviceCode(ctx context.Context, baseURL string) (*deviceCodeResponse, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/device/code", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	var dc deviceCodeResponse
	if err := json.Unmarshal(body, &dc); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &dc, nil
}

// pollDeviceToken returns either a token response (success), an error
// response (pending/slow_down/denied/expired), or a wait hint (seconds).
func pollDeviceToken(ctx context.Context, baseURL, deviceCode string) (*deviceTokenResponse, *deviceTokenErrorResponse, int) {
	payload, _ := json.Marshal(map[string]string{"device_code": deviceCode})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/device/token", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &deviceTokenErrorResponse{Error: err.Error()}, 0
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var t deviceTokenResponse
		if err := json.Unmarshal(body, &t); err != nil {
			return nil, &deviceTokenErrorResponse{Error: fmt.Sprintf("parse response: %v", err)}, 0
		}
		return &t, nil, 0
	}
	var e deviceTokenErrorResponse
	_ = json.Unmarshal(body, &e)
	if e.Error == "" {
		e.Error = fmt.Sprintf("status %d: %s", resp.StatusCode, string(body))
	}
	// honor Retry-After if present
	wait := 0
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		var n int
		_, _ = fmt.Sscanf(ra, "%d", &n)
		wait = n
	}
	return nil, &e, wait
}

// openURL opens the given URL in the user's default browser. No-op on
// unsupported platforms (returns an error the caller may print).
func openURL(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// ExitOnError prints err to stderr and exits, matching the CLI convention.
func ExitOnError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
