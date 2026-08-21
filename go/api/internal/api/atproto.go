package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/fuegoio/sunred/go/api/internal/atproto"
	"github.com/fuegoio/sunred/go/api/internal/auth"
	"github.com/fuegoio/sunred/go/api/internal/store"
)

// --- Input/output types ---

type ATProtoStatusOutput struct {
	Body struct {
		Connected bool   `json:"connected"`
		DID       string `json:"did,omitempty"`
		Handle    string `json:"handle,omitempty"`
	}
}

// registerATProtoRoutes registers AT Protocol integration endpoints.
//
// Identity is now established via OAuth at login (see oauth_handlers.go), so
// there is no app-password "connect" endpoint. The status endpoint reports the
// DID linked to the current user.
func (a *API) registerATProtoRoutes() {
	// GET /.well-known/atproto-did — lets users point an AT Proto handle at
	// this instance. Registered on the bare mux in main.go; here we register
	// the XRPC-namespaced version for discoverability.
	huma.Register(a.huma, huma.Operation{
		OperationID: "atproto-status",
		Method:      http.MethodGet,
		Path:        "/api/v1/me/atproto",
		Summary:     "Get AT Proto identity for the current user",
		Tags:        []string{"social"},
	}, func(ctx context.Context, _ *struct{}) (*ATProtoStatusOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		did, handle, err := a.store.GetUserDID(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		out := &ATProtoStatusOutput{}
		if did != "" {
			out.Body.Connected = true
			out.Body.DID = did
			out.Body.Handle = handle
		}
		return out, nil
	})
}

// --- AT Proto side-effects called from other handlers ---

// atprotoWriterForUser returns an authenticated Writer for userID, or nil if
// the user has no AT Proto connection or credentials are empty.
// If the access token has expired, it attempts a refresh first.
func (a *API) atprotoWriterForUser(ctx context.Context, userID int) (*atproto.Writer, *store.ATProtoCredentials, error) {
	creds, err := a.store.GetATProtoCredentials(ctx, userID)
	if err != nil || creds == nil {
		return nil, nil, err
	}

	// Refresh if the token is expired or will expire within 5 minutes.
	if creds.ExpiresAt != nil && time.Until(*creds.ExpiresAt) < 5*time.Minute {
		refreshClient := atproto.NewClient(creds.PDSUrl, "")
		newSession, err := refreshClient.RefreshSession(ctx, creds.RefreshToken)
		if err != nil {
			slog.Warn("atproto: token refresh failed", "user_id", userID, "err", err)
			// Continue with the stale token — the PDS will return 401 if it's
			// truly expired and the caller will see a write error.
		} else {
			expires := time.Now().Add(2 * time.Hour)
			_ = a.store.UpdateATProtoTokens(ctx, userID, newSession.AccessJwt, newSession.RefreshJwt, &expires)
			creds.AccessToken = newSession.AccessJwt
			creds.RefreshToken = newSession.RefreshJwt
		}
	}

	return atproto.NewWriter(creds.PDSUrl, creds.DID, creds.AccessToken), creds, nil
}

// ATProtoSyncFollow writes or deletes a follow record on the PDS.
// Called fire-and-forget from FollowUser / UnfollowUser handlers.
func (a *API) ATProtoSyncFollow(userID, followeeUserID int, followeeHandle string, isFollow bool) {
	ctx := context.Background()
	w, creds, err := a.atprotoWriterForUser(ctx, userID)
	if err != nil || w == nil {
		slog.Debug("atproto: skip follow sync, no credentials", "user_id", userID, "is_follow", isFollow)
		return
	}

	// Resolve followee DID — they may or may not have connected AT Proto.
	followeeProfile, err := a.store.GetProfileByHandle(ctx, followeeHandle, 0)
	if err != nil || followeeProfile == nil || followeeProfile.DID == "" {
		slog.Debug("atproto: skip follow sync, followee has no DID", "user_id", userID, "handle", followeeHandle)
		return
	}

	if isFollow {
		slog.Info("atproto: writing follow to PDS", "user_id", userID, "subject_did", followeeProfile.DID)
		rkey, err := w.PutFollow(ctx, followeeProfile.DID)
		if err != nil {
			slog.Warn("atproto: put follow failed", "user_id", userID, "err", err)
			return
		}
		slog.Info("atproto: follow written", "user_id", userID, "rkey", rkey)
		_ = a.store.SetFollowATProtoRkey(ctx, userID, followeeUserID, rkey)
	} else {
		rkey, err := a.store.GetFollowATProtoRkey(ctx, userID, followeeUserID)
		if err != nil || rkey == "" {
			return
		}
		slog.Info("atproto: deleting follow from PDS", "user_id", userID, "rkey", rkey)
		if err := w.DeleteFollow(ctx, rkey); err != nil {
			slog.Warn("atproto: delete follow failed", "user_id", userID, "rkey", rkey, "err", err)
		} else {
			slog.Info("atproto: follow deleted", "user_id", userID, "rkey", rkey)
		}
		_ = a.store.SetFollowATProtoRkey(ctx, userID, followeeUserID, "")
	}
	_ = creds
}

// ATProtoSyncShare writes or deletes a share record on the PDS.
func (a *API) ATProtoSyncShare(userID int, sa *store.SharedArticle, isShare bool) {
	ctx := context.Background()
	w, _, err := a.atprotoWriterForUser(ctx, userID)
	if err != nil || w == nil {
		slog.Debug("atproto: skip share sync, no credentials", "user_id", userID, "is_share", isShare)
		return
	}

	if isShare {
		slog.Info("atproto: writing share to PDS", "user_id", userID, "article_url", sa.ArticleURL)
		rkey, err := w.PutShare(ctx,
			sa.ArticleURL, sa.Title, sa.Description,
			sa.FeedURL, sa.FeedTitle, sa.FeedSiteURL,
			sa.Author, sa.PublishedAt, sa.SharedAt,
		)
		if err != nil {
			slog.Warn("atproto: put share failed", "user_id", userID, "err", err)
			return
		}
		slog.Info("atproto: share written", "user_id", userID, "rkey", rkey)
		_ = a.store.SetShareATProtoRkey(ctx, sa.ID, rkey)
	} else {
		rkey, err := a.store.GetShareATProtoRkey(ctx, sa.ID)
		if err != nil || rkey == "" {
			return
		}
		slog.Info("atproto: deleting share from PDS", "user_id", userID, "rkey", rkey)
		if err := w.DeleteShare(ctx, rkey); err != nil {
			slog.Warn("atproto: delete share failed", "user_id", userID, "rkey", rkey, "err", err)
		} else {
			slog.Info("atproto: share deleted", "user_id", userID, "rkey", rkey)
		}
	}
}

// ATProtoSyncFeedSubscription writes or deletes a feed subscription record on
// the PDS. On subscribe it creates an io.sunred.feed.subscription record and
// stores the rkey locally; on unsubscribe it deletes the record using the
// stored rkey.
func (a *API) ATProtoSyncFeedSubscription(userID, feedID int, feedURL, siteURL, title string, isSubscribe bool, createdAt time.Time) {
	ctx := context.Background()
	w, _, err := a.atprotoWriterForUser(ctx, userID)
	if err != nil || w == nil {
		slog.Debug("atproto: skip feed subscription sync, no credentials", "user_id", userID, "is_subscribe", isSubscribe)
		return
	}

	if isSubscribe {
		slog.Info("atproto: writing feed subscription to PDS", "user_id", userID, "feed_url", feedURL)
		rkey, err := w.PutFeedSubscription(ctx, feedURL, siteURL, title, createdAt)
		if err != nil {
			slog.Warn("atproto: put feed subscription failed", "user_id", userID, "feed_url", feedURL, "err", err)
			return
		}
		slog.Info("atproto: feed subscription written", "user_id", userID, "feed_url", feedURL, "rkey", rkey)
		_ = a.store.SetFeedATProtoRkey(ctx, feedID, rkey)
	} else {
		rkey, err := a.store.GetFeedATProtoRkey(ctx, feedID)
		if err != nil || rkey == "" {
			return
		}
		slog.Info("atproto: deleting feed subscription from PDS", "user_id", userID, "feed_url", feedURL, "rkey", rkey)
		if err := w.DeleteFeedSubscription(ctx, rkey); err != nil {
			slog.Warn("atproto: delete feed subscription failed", "user_id", userID, "rkey", rkey, "err", err)
		} else {
			slog.Info("atproto: feed subscription deleted", "user_id", userID, "rkey", rkey)
		}
	}
}

// ATProtoSyncFeedList writes or deletes a feed list record on the PDS.
func (a *API) ATProtoSyncFeedList(userID, listID int, fl *store.FeedList, isCreate bool) {
	ctx := context.Background()
	w, _, err := a.atprotoWriterForUser(ctx, userID)
	if err != nil || w == nil {
		slog.Debug("atproto: skip feed list sync, no credentials", "user_id", userID, "is_create", isCreate)
		return
	}

	if !isCreate {
		rkey, err := a.store.GetFeedListATProtoRkey(ctx, listID)
		if err != nil || rkey == "" {
			return
		}
		slog.Info("atproto: deleting feed list from PDS", "user_id", userID, "rkey", rkey)
		if err := w.DeleteFeedList(ctx, rkey); err != nil {
			slog.Warn("atproto: delete feed list failed", "user_id", userID, "rkey", rkey, "err", err)
		} else {
			slog.Info("atproto: feed list deleted", "user_id", userID, "rkey", rkey)
		}
		return
	}

	// Build the feed entry list from the full FeedList.
	entries := make([]atproto.FeedListEntry, 0, len(fl.Feeds))
	for _, f := range fl.Feeds {
		entries = append(entries, atproto.FeedListEntry{
			FeedURL: f.FeedURL,
			SiteURL: f.SiteURL,
			Title:   f.Title,
		})
	}

	existingRkey, _ := a.store.GetFeedListATProtoRkey(ctx, listID)
	slog.Info("atproto: writing feed list to PDS", "user_id", userID, "list_id", listID, "rkey", existingRkey)
	rkey, err := w.PutFeedList(ctx, existingRkey, fl.Title, fl.Description, fl.IsPublic, entries, fl.CreatedAt)
	if err != nil {
		slog.Warn("atproto: put feed list failed", "user_id", userID, "err", err)
		return
	}
	slog.Info("atproto: feed list written", "user_id", userID, "rkey", rkey)
	_ = a.store.SetFeedListATProtoRkey(ctx, listID, rkey)
}

// WellKnownATProtoDIDHandler returns an http.Handler for /.well-known/atproto-did.
// It resolves the host's subdomain (or ?handle= query param) to a DID.
func (a *API) WellKnownATProtoDIDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract handle from subdomain: "fuego.sunred.example" → "fuego"
		host := r.Host
		handle := r.URL.Query().Get("handle")
		if handle == "" {
			// Strip port from host
			for i := len(host) - 1; i >= 0; i-- {
				if host[i] == ':' {
					host = host[:i]
					break
				}
			}
			// First segment before the first dot is the handle subdomain
			if idx := strings.IndexByte(host, '.'); idx > 0 {
				handle = host[:idx]
			}
		}
		if handle == "" {
			http.Error(w, "handle not found", http.StatusNotFound)
			return
		}

		profile, err := a.store.GetProfileByHandle(r.Context(), handle, 0)
		if err != nil || profile == nil || profile.DID == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(profile.DID))
	})
}
