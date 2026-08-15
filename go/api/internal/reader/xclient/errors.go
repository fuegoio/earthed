package xclient

import "errors"

// ErrNotConfigured is returned when the X API client has no bearer token.
var ErrNotConfigured = errors.New("xclient: not configured (X_API_BEARER_TOKEN is unset)")
