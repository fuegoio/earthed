package atproto

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
)

// NewOAuthApp builds the indigo OAuth ClientApp for Sunred.
//
// For local development (BaseURL on 127.0.0.1 or localhost), it uses
// NewLocalhostConfig, which encodes the client metadata directly in the
// client_id as query parameters — the PDS accepts this without fetching a
// metadata document (which it couldn't reach on a loopback address anyway).
//
// For production (https BaseURL), it uses NewPublicConfig with a real
// client_id URL pointing at the client-metadata.json document this server
// serves.
//
// Sunred is a public (non-confidential) client: it has no shared secret with
// the PDS, relying on PKCE + DPoP instead. The PGStore persists auth-request
// state and sessions in PostgreSQL so logins survive restarts.
func NewOAuthApp(db *sql.DB, clientID, callbackURL string) (*oauth.ClientApp, error) {
	if clientID == "" || callbackURL == "" {
		return nil, fmt.Errorf("oauth client_id and callback URL are required")
	}

	// "atproto" is the base scope; "transition:email" lets us read the
	// account email via com.atproto.server.getSession on callback.
	scopes := []string{"atproto", "transition:email"}

	var config oauth.ClientConfig
	if isLoopbackURL(callbackURL) {
		config = oauth.NewLocalhostConfig(callbackURL, scopes)
	} else {
		config = oauth.NewPublicConfig(clientID, callbackURL, scopes)
	}

	store := NewPGStore(db)
	app := oauth.NewClientApp(&config, store)
	return app, nil
}

// isLoopbackURL reports whether the URL points at a loopback address
// (127.0.0.1 or localhost), in which case the localhost OAuth client config
// is used.
func isLoopbackURL(rawURL string) bool {
	return strings.Contains(rawURL, "://127.0.0.1") ||
		strings.Contains(rawURL, "://localhost")
}
