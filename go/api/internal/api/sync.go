package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/syntax"

	"github.com/fuegoio/earthed/go/api/internal/atproto"
	"github.com/fuegoio/earthed/go/api/internal/store"
)

// listRecordsOut matches com.atproto.repo.listRecords response.
type listRecordsOut struct {
	Records []struct {
		URI   string          `json:"uri"`
		CID   string          `json:"cid"`
		Value json.RawMessage `json:"value"`
	} `json:"records"`
	Cursor string `json:"cursor"`
}

// syncFollows backfills io.earthed.graph.follow records from the PDS into the
// local follow cache. Each follow record's `subject` is a DID; we record a
// local follow edge if the followee is also a known Earthed user on this
// instance, and store the rkey for later delete-on-unfollow.
func syncFollows(ctx context.Context, c *atclient.APIClient, st *store.Store, userID int) error {
	cursor := ""
	for {
		params := map[string]any{
			"repo":       accountDID(c),
			"collection": atproto.CollectionFollow,
			"limit":       100,
		}
		if cursor != "" {
			params["cursor"] = cursor
		}
		var out listRecordsOut
		if err := c.Get(ctx, "com.atproto.repo.listRecords", params, &out); err != nil {
			return fmt.Errorf("list follows: %w", err)
		}
		for _, rec := range out.Records {
			var f struct {
				Subject   string `json:"subject"`
				CreatedAt string `json:"createdAt"`
			}
			if err := json.Unmarshal(rec.Value, &f); err != nil || f.Subject == "" {
				continue
			}
			rkey := rkeyFromURI(rec.URI)
			// Record the follow edge locally if the subject is a local user.
			if followeeID, _ := st.GetUserIDByDID(ctx, f.Subject); followeeID != 0 {
				_ = st.UpsertFollowWithRkey(ctx, userID, followeeID, rkey)
			}
		}
		if out.Cursor == "" {
			break
		}
		cursor = out.Cursor
	}
	return nil
}

// syncFeedSubscriptions backfills io.earthed.feed.subscription records into the
// local feeds table, storing the rkey so later unsubscribe deletes the record.
func syncFeedSubscriptions(ctx context.Context, c *atclient.APIClient, st *store.Store, userID int) error {
	cursor := ""
	for {
		params := map[string]any{
			"repo":       accountDID(c),
			"collection": atproto.CollectionSubscription,
			"limit":       100,
		}
		if cursor != "" {
			params["cursor"] = cursor
		}
		var out listRecordsOut
		if err := c.Get(ctx, "com.atproto.repo.listRecords", params, &out); err != nil {
			return fmt.Errorf("list feed subs: %w", err)
		}
		for _, rec := range out.Records {
			var fs struct {
				FeedURL    string `json:"feedUrl"`
				SiteURL    string `json:"siteUrl"`
				Title      string `json:"title"`
				CreatedAt  string `json:"createdAt"`
			}
			if err := json.Unmarshal(rec.Value, &fs); err != nil || fs.FeedURL == "" {
				continue
			}
			rkey := rkeyFromURI(rec.URI)
			if err := st.UpsertFeedSubscriptionWithRkey(ctx, userID, fs.FeedURL, fs.SiteURL, fs.Title, rkey); err != nil {
				slog.Warn("sync: upsert feed sub", "feed_url", fs.FeedURL, "err", err)
			}
		}
		if out.Cursor == "" {
			break
		}
		cursor = out.Cursor
	}
	return nil
}

// accountDID returns the DID string the API client is authenticated as.
func accountDID(c *atclient.APIClient) string {
	if c == nil || c.AccountDID == nil {
		return ""
	}
	return string(*c.AccountDID)
}

// rkeyFromURI extracts the record key from an at:// URI (at://did/collection/rkey).
func rkeyFromURI(uri string) string {
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			return uri[i+1:]
		}
	}
	return uri
}

// keep syntax import referenced (DID parsing helper for future use)
var _ = syntax.DID("")
