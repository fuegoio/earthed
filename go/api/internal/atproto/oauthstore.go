package atproto

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// PGStore implements oauth.ClientAuthStore backed by PostgreSQL. It persists
// in-flight auth requests and OAuth sessions as JSONB so the server survives
// restarts and supports concurrent sessions per DID.
type PGStore struct {
	DB *sql.DB
}

// NewPGStore returns a Postgres-backed OAuth client store.
func NewPGStore(db *sql.DB) *PGStore {
	return &PGStore{DB: db}
}

// --- Auth request info ---

func (s *PGStore) GetAuthRequestInfo(ctx context.Context, state string) (*oauth.AuthRequestData, error) {
	var raw []byte
	err := s.DB.QueryRowContext(ctx,
		`SELECT data FROM oauth_auth_requests WHERE state = $1`, state,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("auth request not found")
	}
	if err != nil {
		return nil, err
	}
	var info oauth.AuthRequestData
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("decode auth request: %w", err)
	}
	return &info, nil
}

func (s *PGStore) SaveAuthRequestInfo(ctx context.Context, info oauth.AuthRequestData) error {
	raw, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("encode auth request: %w", err)
	}
	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO oauth_auth_requests (state, data) VALUES ($1, $2) ON CONFLICT (state) DO NOTHING`,
		info.State, raw,
	)
	return err
}

func (s *PGStore) DeleteAuthRequestInfo(ctx context.Context, state string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM oauth_auth_requests WHERE state = $1`, state)
	return err
}

// --- Sessions ---

func (s *PGStore) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*oauth.ClientSessionData, error) {
	var raw []byte
	err := s.DB.QueryRowContext(ctx,
		`SELECT data FROM oauth_sessions WHERE account_did = $1 AND session_id = $2`,
		string(did), sessionID,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}
	var sess oauth.ClientSessionData
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	return &sess, nil
}

func (s *PGStore) SaveSession(ctx context.Context, sess oauth.ClientSessionData) error {
	raw, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO oauth_sessions (account_did, session_id, data, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (account_did, session_id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`,
		string(sess.AccountDID), sess.SessionID, raw,
	)
	return err
}

func (s *PGStore) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	_, err := s.DB.ExecContext(ctx,
		`DELETE FROM oauth_sessions WHERE account_did = $1 AND session_id = $2`,
		string(did), sessionID,
	)
	return err
}
