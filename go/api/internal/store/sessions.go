package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// --- Web sessions (cookie auth) ---

// CreateWebSession inserts a new web session token for a user.
func (s *Store) CreateWebSession(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO web_sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	return err
}

// GetWebSession resolves a session token to a user ID. Returns 0 if the token
// is unknown or expired.
func (s *Store) GetWebSession(ctx context.Context, token string) (int, error) {
	if token == "" {
		return 0, fmt.Errorf("empty session token")
	}
	var userID int
	err := s.DB.QueryRowContext(ctx,
		`SELECT user_id FROM web_sessions WHERE token = $1 AND expires_at > NOW()`, token,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid session")
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteWebSession removes a web session token.
func (s *Store) DeleteWebSession(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM web_sessions WHERE token = $1`, token)
	return err
}

// --- DID-based users ---

// UpdateUserEmail sets the email on a user row. Called after OAuth callback
// when the PDS returns the account email via com.atproto.server.getSession
// (requires the transition:email scope).
func (s *Store) UpdateUserEmail(ctx context.Context, userID int, email string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE users SET email = $2 WHERE id = $1`,
		userID, email,
	)
	return err
}

// GetOrCreateUserByDID returns the user ID for a DID, creating the user (and a
// blank profile row) if it does not yet exist. handle is stored on the users
// row; a user_profiles row is created so social queries join correctly.
func (s *Store) GetOrCreateUserByDID(ctx context.Context, did, handle string) (int, bool, error) {
	var userID int
	err := s.DB.QueryRowContext(ctx,
		`SELECT id FROM users WHERE did = $1`, did,
	).Scan(&userID)
	if err == nil {
		// Existing user: refresh the handle in case it changed on the PDS.
		if handle != "" {
			_, _ = s.DB.ExecContext(ctx, `UPDATE users SET handle = $2 WHERE id = $1 AND $2 <> COALESCE(handle, '')`, userID, handle)
		}
		return userID, false, nil
	}
	if err != sql.ErrNoRows {
		return 0, false, fmt.Errorf("lookup user by did: %w", err)
	}

	created := false
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO users (did, handle) VALUES ($1, $2)
		 ON CONFLICT (did) WHERE did IS NOT NULL DO UPDATE SET handle = EXCLUDED.handle
		 RETURNING id, (xmax = 0)`,
		did, handle,
	).Scan(&userID, &created)
	if err != nil {
		return 0, false, fmt.Errorf("create user by did: %w", err)
	}
	// Ensure a user_profiles row exists for social joins.
	_, _ = s.DB.ExecContext(ctx,
		`INSERT INTO user_profiles (user_id, handle, bio) VALUES ($1, $2, '')
		 ON CONFLICT (user_id) DO NOTHING`,
		userID, handle,
	)
	return userID, created, nil
}

// --- Sync helpers (PDS → local cache) ---

// GetUserIDByDID resolves a DID to a local user ID, or 0 if unknown.
func (s *Store) GetUserIDByDID(ctx context.Context, did string) (int, error) {
	var id int
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE did = $1`, did).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// GetUserDID returns the DID and handle for a user, or empty strings if unset.
func (s *Store) GetUserDID(ctx context.Context, userID int) (did, handle string, err error) {
	var d, h sql.NullString
	err = s.DB.QueryRowContext(ctx,
		`SELECT did, handle FROM users WHERE id = $1`, userID,
	).Scan(&d, &h)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return d.String, h.String, err
}

// UpsertFollowWithRkey records a local follow edge and its AT Proto rkey.
func (s *Store) UpsertFollowWithRkey(ctx context.Context, followerID, followeeID int, rkey string) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO user_follows (follower_id, followee_id, atproto_rkey)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (follower_id, followee_id) DO UPDATE SET atproto_rkey = EXCLUDED.atproto_rkey`,
		followerID, followeeID, rkey,
	)
	return err
}

// UpsertFeedSubscriptionWithRkey records a feed subscription with its AT Proto
// rkey so a later unsubscribe can delete the record.
func (s *Store) UpsertFeedSubscriptionWithRkey(ctx context.Context, userID int, feedURL, siteURL, title, rkey string) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO feeds (user_id, feed_url, site_url, title, atproto_rkey)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, feed_url) DO UPDATE SET
		   site_url = EXCLUDED.site_url,
		   title = EXCLUDED.title,
		   atproto_rkey = EXCLUDED.atproto_rkey,
		   updated_at = NOW()`,
		userID, feedURL, siteURL, title, rkey,
	)
	return err
}
