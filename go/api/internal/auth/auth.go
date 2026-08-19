// Package auth wires the Limen identity provider to the Earthed HTTP API
// and exposes session and bearer-token middleware.
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/thecodearcher/limen"
	sqladapter "github.com/thecodearcher/limen/adapters/sql"
	credentialpassword "github.com/thecodearcher/limen/plugins/credential-password"

	"github.com/fuegoio/earthed/go/api/internal/config"
	"github.com/fuegoio/earthed/go/api/internal/store"
)

type contextKey int

const userKey contextKey = iota

// Auth bundles the Limen instance with the Earthed store and database.
type Auth struct {
	Limen *limen.Limen
	Store *store.Store
	DB    *sql.DB
}

// New builds an Auth instance configured from cfg, attaching the password
// credential plugin. Email verification and password-reset callbacks log
// instead of sending mail in this scaffold; wire a real mailer when needed.
func New(cfg *config.Config, db *sql.DB, st *store.Store) (*Auth, error) {
	plugins := []limen.Plugin{
		credentialpassword.New(
			credentialpassword.WithSendPasswordResetEmail(func(email, token string) {
				slog.Info("password reset requested", "email", email, "token", token)
			}),
		),
	}

	authInstance, err := limen.New(&limen.Config{
		BaseURL:  cfg.BaseURL,
		Database: sqladapter.NewPostgreSQL(db),
		Secret:   []byte(cfg.LimenSecret),
		HTTP: limen.NewDefaultHTTPConfig(
			limen.WithHTTPBasePath("/auth"),
			limen.WithHTTPCSRFProtection(false),
			limen.WithHTTPOriginCheck(false),
			limen.WithHTTPCookieSecure(false),
		),
		Email: limen.NewDefaultEmailConfig(
			limen.WithEmailVerification(
				limen.WithSendEmailVerificationMail(func(email, token string) {
					slog.Info("email verification requested", "email", email, "token", token)
				}),
			),
		),
		Plugins: plugins,
	})
	if err != nil {
		return nil, fmt.Errorf("init limen: %w", err)
	}

	return &Auth{
		Limen: authInstance,
		Store: st,
		DB:    db,
	}, nil
}

// Handler returns the http.Handler that serves Limen auth endpoints.
func (a *Auth) Handler() http.Handler {
	return a.Limen.Handler()
}

// Middleware wraps protected API routes. It verifies the Limen session
// (or bearer API token) and injects the authenticated user ID into the
// request context.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) resolveUserID(r *http.Request) (int, error) {
	if bearer := bearerToken(r); bearer != "" {
		return a.resolveToken(r.Context(), bearer)
	}
	session, err := a.Limen.GetSession(r)
	if err != nil {
		return 0, err
	}
	id, err := IDToInt(session.User.ID)
	return id, err
}

func (a *Auth) resolveToken(ctx context.Context, token string) (int, error) {
	hash := HashToken(token)
	t, err := a.Store.GetAPITokenByHash(ctx, hash)
	if err != nil || t == nil {
		return 0, fmt.Errorf("invalid token")
	}
	return t.UserID, nil
}

// UserIDFromCtx extracts the authenticated user id stored by Middleware, or 0.
func UserIDFromCtx(ctx context.Context) int {
	v, _ := ctx.Value(userKey).(int)
	return v
}

// IDToInt coerces a Limen user id (int, int64 or float64) into an int.
func IDToInt(id any) (int, error) {
	switch v := id.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	}
	return 0, fmt.Errorf("unexpected user id type: %T", id)
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
