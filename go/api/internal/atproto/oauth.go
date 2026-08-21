package atproto

import (
	"database/sql"
	"fmt"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
)

// NewOAuthApp builds the indigo OAuth ClientApp for Earthed.
//
// clientID is the full URL where Earthed serves its client metadata document
// (EARTHED_OAUTH_CLIENT_ID, typically "<BaseURL>/client-metadata.json").
// callbackURL is the public OAuth callback URL.
//
// Earthed is a public (non-confidential) client: it has no shared secret with
// the PDS, relying on PKCE + DPoP instead. The PGStore persists auth-request
// state and sessions in PostgreSQL so logins survive restarts.
func NewOAuthApp(db *sql.DB, clientID, callbackURL string) (*oauth.ClientApp, error) {
	if clientID == "" || callbackURL == "" {
		return nil, fmt.Errorf("oauth client_id and callback URL are required")
	}
	config := oauth.NewPublicConfig(clientID, callbackURL, []string{"atproto"})
	store := NewPGStore(db)
	app := oauth.NewClientApp(&config, store)
	return app, nil
}
