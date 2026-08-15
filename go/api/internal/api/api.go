// Package api registers the Earthed REST API on a huma router.
package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/fuegoio/earthed/go/api/internal/auth"
	"github.com/fuegoio/earthed/go/api/internal/config"
	"github.com/fuegoio/earthed/go/api/internal/reader/discoverer"
	"github.com/fuegoio/earthed/go/api/internal/reader/fetcher"
	"github.com/fuegoio/earthed/go/api/internal/reader/parser"
	"github.com/fuegoio/earthed/go/api/internal/reader/processor"
	"github.com/fuegoio/earthed/go/api/internal/reader/sanitizer"
	"github.com/fuegoio/earthed/go/api/internal/store"
	"github.com/fuegoio/earthed/go/api/internal/reader/xclient"
)

// API wires the store and auth to a huma router and registers the REST routes.
type API struct {
	huma      huma.API
	store     *store.Store
	auth      *auth.Auth
	cfg       *config.Config
	fetcher   *fetcher.Fetcher
	xclient   *xclient.Client
	processor *processor.Processor
}

// New returns an API bound to the given huma router, store, auth, config,
// New returns an API bound to the given huma router, store, auth, config,
// fetcher, and X API client. The fetcher and xclient may be nil when only
// generating the OpenAPI spec (--openapi flag).
func New(humaAPI huma.API, st *store.Store, authInst *auth.Auth, cfg *config.Config, f *fetcher.Fetcher, xc *xclient.Client) *API {
	var proc *processor.Processor
	if st != nil && (f != nil || (xc != nil && xc.Enabled())) {
		proc = processor.New(st, f, xc)
	}
	return &API{huma: humaAPI, store: st, auth: authInst, cfg: cfg, fetcher: f, xclient: xc, processor: proc}
}

// OpenAPITags returns the ordered tag list for the OpenAPI spec.
func OpenAPITags() []*huma.Tag {
	return []*huma.Tag{
		{Name: "feeds", Description: "Feed subscriptions"},
		{Name: "entries", Description: "Feed entries/articles"},
		{Name: "folders", Description: "Feed folders"},
		{Name: "users", Description: "User accounts"},
		{Name: "tokens", Description: "API tokens"},
		{Name: "opml", Description: "OPML import/export"},
		{Name: "feed-lists", Description: "Shareable feed list collections"},
		{Name: "device", Description: "Device-flow login (CLI/TUI)"},
		{Name: "x", Description: "X (Twitter) timeline subscriptions"},
	}
}

// RegisterRoutes registers all Earthed REST routes on the huma router.
func (a *API) RegisterRoutes() {
	a.registerHealthRoutes()
	a.registerMeRoutes()
	a.registerFolderRoutes()
	a.registerFeedRoutes()
	a.registerEntryRoutes()
	a.registerTokenRoutes()
	a.registerFeedListRoutes()
	a.registerXRoutes()
	a.registerOPMLRoutes()
	a.registerDeviceRoutes()
}

// --- Health ---

func (a *API) registerHealthRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check",
		Tags:        []string{"health"},
	}, func(ctx context.Context, _ *struct{}) (*struct {
		Body struct {
			Status string `json:"status"`
		}
	}, error) {
		return &struct {
			Body struct {
				Status string `json:"status"`
			}
		}{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})
}

// --- Me ---

type MeOutput struct {
	Body store.User
}

type UpdateMeInput struct {
	Body struct {
		FirstName string `json:"first_name" maxLength:"255"`
		Email     string `json:"email" format:"email" minLength:"1" maxLength:"255"`
	}
}

func (a *API) registerMeRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/v1/me",
		Summary:     "Get current user",
		Tags:        []string{"users"},
	}, func(ctx context.Context, _ *struct{}) (*MeOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		user, err := a.store.GetUserByID(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Errorf("get user: %w", err).Error())
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}
		return &MeOutput{Body: *user}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "update-me",
		Method:      http.MethodPatch,
		Path:        "/api/v1/me",
		Summary:     "Update current user profile",
		Tags:        []string{"users"},
	}, func(ctx context.Context, input *UpdateMeInput) (*MeOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		user, err := a.store.UpdateUser(ctx, userID, input.Body.FirstName, input.Body.Email)
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Errorf("update user: %w", err).Error())
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}
		return &MeOutput{Body: *user}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-me",
		Method:      http.MethodDelete,
		Path:        "/api/v1/me",
		Summary:     "Delete current user account",
		Tags:        []string{"users"},
	}, func(ctx context.Context, _ *struct{}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteUser(ctx, userID); err != nil {
			return nil, huma.Error500InternalServerError(fmt.Errorf("delete user: %w", err).Error())
		}
		return nil, nil
	})
}

// --- Folders ---

type CreateFolderInput struct {
	Body struct {
		Title    string `json:"title" minLength:"1" maxLength:"255"`
		ParentID *int   `json:"parent_id,omitempty"`
	}
}

type UpdateFolderInput struct {
	FolderID int `path:"folderId"`
	Body     struct {
		Title    string `json:"title" minLength:"1" maxLength:"255"`
		ParentID *int   `json:"parent_id,omitempty"`
	}
}

type FolderOutput struct {
	Body store.Folder
}

type FolderListOutput struct {
	Body []store.Folder
}

func (a *API) registerFolderRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-folder",
		Method:      http.MethodPost,
		Path:        "/api/v1/folders",
		Summary:     "Create a folder",
		Tags:        []string{"folders"},
	}, func(ctx context.Context, input *CreateFolderInput) (*FolderOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		folder, err := a.store.CreateFolder(ctx, userID, input.Body.Title, input.Body.ParentID)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &FolderOutput{Body: *folder}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-folders",
		Method:      http.MethodGet,
		Path:        "/api/v1/folders",
		Summary:     "List folders",
		Tags:        []string{"folders"},
	}, func(ctx context.Context, _ *struct{}) (*FolderListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		folders, err := a.store.ListFolders(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if folders == nil {
			folders = []store.Folder{}
		}
		return &FolderListOutput{Body: folders}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "update-folder",
		Method:      http.MethodPatch,
		Path:        "/api/v1/folders/{folderId}",
		Summary:     "Update a folder",
		Description: "Update the title and/or parent folder of a folder. Set parent_id to move or nest the folder.",
		Tags:        []string{"folders"},
	}, func(ctx context.Context, input *UpdateFolderInput) (*FolderOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		folder, err := a.store.UpdateFolder(ctx, input.FolderID, userID, input.Body.Title, input.Body.ParentID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if folder == nil {
			return nil, huma.Error404NotFound("folder not found")
		}
		return &FolderOutput{Body: *folder}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-folder",
		Method:      http.MethodDelete,
		Path:        "/api/v1/folders/{folderId}",
		Summary:     "Delete a folder",
		Tags:        []string{"folders"},
	}, func(ctx context.Context, input *struct {
		FolderID int `path:"folderId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteFolder(ctx, input.FolderID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})
}

// --- Feeds ---

type CreateFeedInput struct {
	Body struct {
		FeedURL  string `json:"feed_url" minLength:"1" maxLength:"2048"`
		FolderID *int   `json:"folder_id,omitempty"`
	}
}

type UpdateFeedInput struct {
	FeedID int `path:"feedId"`
	Body   struct {
		FolderID *int   `json:"folder_id,omitempty"`
		Title    string `json:"title,omitempty" maxLength:"512"`
	}
}

type FeedOutput struct {
	Body store.Feed
}

type FeedListOutput struct {
	Body []store.Feed
}

type PreviewFeedInput struct {
	Body struct {
		FeedURL string `json:"feed_url" minLength:"1" maxLength:"2048"`
	}
}

type PreviewFeedItem struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author,omitempty"`
	Description string    `json:"description,omitempty"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"published_at"`
	Tags        []string  `json:"tags,omitempty"`
}

type PreviewFeedBody struct {
	Title       string            `json:"title"`
	SiteURL     string            `json:"site_url"`
	FeedURL     string            `json:"feed_url"`
	Description string            `json:"description,omitempty"`
	FaviconURL  string            `json:"favicon_url,omitempty"`
	Items       []PreviewFeedItem `json:"items"`
}

type PreviewFeedOutput struct {
	Body PreviewFeedBody
}

func (a *API) registerFeedRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds",
		Summary:     "Subscribe to a feed",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *CreateFeedInput) (*FeedOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		feed, err := a.subscribeToFeed(ctx, userID, input.Body.FeedURL, input.Body.FolderID)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &FeedOutput{Body: *feed}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-feeds",
		Method:      http.MethodGet,
		Path:        "/api/v1/feeds",
		Summary:     "List feeds",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, _ *struct{}) (*FeedListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		feeds, err := a.store.ListFeeds(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feeds == nil {
			feeds = []store.Feed{}
		}
		return &FeedListOutput{Body: feeds}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "get-feed",
		Method:      http.MethodGet,
		Path:        "/api/v1/feeds/{feedId}",
		Summary:     "Get a feed",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *struct {
		FeedID int `path:"feedId"`
	}) (*FeedOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		feed, err := a.store.GetFeedByID(ctx, input.FeedID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feed == nil {
			return nil, huma.Error404NotFound("feed not found")
		}
		return &FeedOutput{Body: *feed}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-feed",
		Method:      http.MethodDelete,
		Path:        "/api/v1/feeds/{feedId}",
		Summary:     "Delete a feed",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *struct {
		FeedID int `path:"feedId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteFeed(ctx, input.FeedID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "update-feed",
		Method:      http.MethodPatch,
		Path:        "/api/v1/feeds/{feedId}",
		Summary:     "Update a feed",
		Description: "Update the folder assignment and/or title of a feed.",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *UpdateFeedInput) (*FeedOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		feed, err := a.store.GetFeedByID(ctx, input.FeedID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feed == nil {
			return nil, huma.Error404NotFound("feed not found")
		}
		folderID := input.Body.FolderID
		if folderID == nil {
			folderID = feed.FolderID
		}
		title := input.Body.Title
		if title == "" {
			title = feed.Title
		}
		updated, err := a.store.UpdateFeed(ctx, input.FeedID, userID, folderID, title, feed.ScraperRules, feed.RewriteRules, feed.Disabled, feed.Crawler)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if updated == nil {
			return nil, huma.Error404NotFound("feed not found")
		}
		return &FeedOutput{Body: *updated}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "mark-feed-read",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds/{feedId}/mark-all-read",
		Summary:     "Mark all entries in a feed as read",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *struct {
		FeedID int `path:"feedId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.MarkFeedEntriesRead(ctx, input.FeedID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "refresh-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds/{feedId}/refresh",
		Summary:     "Refresh a feed",
		Description: "Manually fetch and parse the feed, inserting any new entries. Use this to get the latest articles without waiting for the scheduler.",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *struct {
		FeedID int `path:"feedId"`
	}) (*FeedOutput, error) {
		if a.processor == nil {
			return nil, huma.Error503ServiceUnavailable("feed processor is not available")
		}
		userID := auth.UserIDFromCtx(ctx)
		feed, err := a.store.GetFeedByID(ctx, input.FeedID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feed == nil {
			return nil, huma.Error404NotFound("feed not found")
		}
		if err := a.processor.ProcessFeed(ctx, feed); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		feed, err = a.store.GetFeedByID(ctx, input.FeedID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feed == nil {
			return nil, huma.Error404NotFound("feed not found")
		}
		return &FeedOutput{Body: *feed}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "preview-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds/preview",
		Summary:     "Preview a feed without subscribing",
		Description: "Fetches and parses a feed URL, returning feed metadata and recent entries without persisting anything.",
		Tags:        []string{"feeds"},
	}, func(ctx context.Context, input *PreviewFeedInput) (*PreviewFeedOutput, error) {
		if a.fetcher == nil {
			return nil, huma.Error503ServiceUnavailable("feed fetcher is not available")
		}

		// Try to discover the feed URL. If the input is already a feed URL,
		// discovery returns it as-is. If the input is an HTML page, discovery
		// parses <link rel="alternate"> tags to find the feed URL.
		discovery, err := discoverer.Discover(ctx, a.fetcher, input.Body.FeedURL)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("could not discover feed: %s", err.Error()))
		}

		feedURL := discovery.FeedURL
		result, err := a.fetcher.Fetch(ctx, feedURL, "", "")
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("could not fetch feed: %s", err.Error()))
		}
		if result.NotModified {
			return nil, huma.Error400BadRequest("feed returned 304 Not Modified — no content to preview")
		}

		parsed, err := parser.Parse(result.Body, result.ContentType)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("could not parse feed: %s", err.Error()))
		}

		maxItems := 20
		if len(parsed.Items) > maxItems {
			parsed.Items = parsed.Items[:maxItems]
		}

		items := make([]PreviewFeedItem, 0, len(parsed.Items))
		for _, item := range parsed.Items {
			content := item.Content
			if content == "" {
				content = item.Description
			}
			sanitized, err := sanitizer.Sanitize(content)
			if err != nil {
				sanitized = content
			}

			publishedAt, err := processor.ParseDate(item.PublishedAt)
			if err != nil {
				publishedAt = time.Now()
			}

			items = append(items, PreviewFeedItem{
				Title:       item.Title,
				URL:         item.Link,
				Author:      item.Author,
				Description: sanitizer.StripHTML(item.Description),
				Content:     sanitized,
				PublishedAt: publishedAt,
				Tags:        item.Tags,
			})
		}

		faviconURL := ""
		if parsed.SiteURL != "" {
			faviconURL = "https://www.google.com/s2/favicons?domain=" + parsed.SiteURL + "&sz=64"
		}

		return &PreviewFeedOutput{Body: PreviewFeedBody{
			Title:       parsed.Title,
			SiteURL:     parsed.SiteURL,
			FeedURL:     feedURL,
			Description: parsed.Description,
			FaviconURL:  faviconURL,
			Items:       items,
		}}, nil
	})
}

// --- Entries ---

type EntryListOutput struct {
	Body []store.Entry
}

type EntryOutput struct {
	Body store.Entry
}

func (a *API) registerEntryRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "list-entries",
		Method:      http.MethodGet,
		Path:        "/api/v1/entries",
		Summary:     "List entries",
		Tags:        []string{"entries"},
	}, func(ctx context.Context, input *struct {
		FeedID   int    `query:"feed_id" omitempty:""`
		FolderID int    `query:"folder_id" omitempty:""`
		Status   string `query:"status" enum:"unread,read,removed" omitempty:""`
		Starred  bool   `query:"starred" omitempty:""`
		Search   string `query:"search" omitempty:""`
		Limit    int    `query:"limit" default:"50" maximum:"200"`
		Offset   int    `query:"offset" default:"0"`
	}) (*EntryListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		if input.Limit == 0 {
			input.Limit = 50
		}
		var feedID *int
		if input.FeedID > 0 {
			feedID = &input.FeedID
		}
		var folderID *int
		if input.FolderID > 0 {
			folderID = &input.FolderID
		}
		var starred *bool
		if input.Starred {
			starred = &input.Starred
		}
		entries, err := a.store.ListEntries(ctx, userID, feedID, folderID, input.Status, starred, input.Search, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if entries == nil {
			entries = []store.Entry{}
		}
		return &EntryListOutput{Body: entries}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "get-entry",
		Method:      http.MethodGet,
		Path:        "/api/v1/entries/{entryId}",
		Summary:     "Get an entry",
		Tags:        []string{"entries"},
	}, func(ctx context.Context, input *struct {
		EntryID int64 `path:"entryId"`
	}) (*EntryOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		entry, err := a.store.GetEntryByID(ctx, input.EntryID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if entry == nil {
			return nil, huma.Error404NotFound("entry not found")
		}
		return &EntryOutput{Body: *entry}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "update-entries",
		Method:      http.MethodPut,
		Path:        "/api/v1/entries",
		Summary:     "Bulk update entry status",
		Tags:        []string{"entries"},
	}, func(ctx context.Context, input *struct {
		Body struct {
			EntryIDs []int64 `json:"entry_ids"`
			Status   string  `json:"status" enum:"unread,read,removed"`
		}
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.UpdateEntryStatus(ctx, input.Body.EntryIDs, userID, input.Body.Status); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "toggle-entry-starred",
		Method:      http.MethodPut,
		Path:        "/api/v1/entries/{entryId}/starred",
		Summary:     "Toggle starred on an entry",
		Tags:        []string{"entries"},
	}, func(ctx context.Context, input *struct {
		EntryID int64 `path:"entryId"`
		Body    struct {
			Starred bool `json:"starred"`
		}
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.ToggleEntryStarred(ctx, input.EntryID, userID, input.Body.Starred); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})
}

// --- API Tokens ---

type CreateTokenInput struct {
	Body struct {
		Label string `json:"label" minLength:"1" maxLength:"255"`
	}
}

type TokenOutput struct {
	Body struct {
		ID        int       `json:"id"`
		Label     string    `json:"label"`
		Token     string    `json:"token"`
		CreatedAt time.Time `json:"created_at"`
	}
}

type TokenListOutput struct {
	Body []store.APIToken
}

func (a *API) registerTokenRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-token",
		Method:      http.MethodPost,
		Path:        "/api/v1/tokens",
		Summary:     "Create an API token",
		Tags:        []string{"tokens"},
	}, func(ctx context.Context, input *CreateTokenInput) (*TokenOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		rawToken := generateToken()
		hash := auth.HashToken(rawToken)
		t, err := a.store.CreateAPIToken(ctx, userID, input.Body.Label, hash, "manual", nil)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &TokenOutput{
			Body: struct {
				ID        int       `json:"id"`
				Label     string    `json:"label"`
				Token     string    `json:"token"`
				CreatedAt time.Time `json:"created_at"`
			}{
				ID:        t.ID,
				Label:     t.Label,
				Token:     rawToken,
				CreatedAt: t.CreatedAt,
			},
		}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-tokens",
		Method:      http.MethodGet,
		Path:        "/api/v1/tokens",
		Summary:     "List API tokens",
		Tags:        []string{"tokens"},
	}, func(ctx context.Context, _ *struct{}) (*TokenListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		tokens, err := a.store.ListAPITokens(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if tokens == nil {
			tokens = []store.APIToken{}
		}
		return &TokenListOutput{Body: tokens}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-token",
		Method:      http.MethodDelete,
		Path:        "/api/v1/tokens/{tokenId}",
		Summary:     "Delete an API token",
		Tags:        []string{"tokens"},
	}, func(ctx context.Context, input *struct {
		TokenID int `path:"tokenId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteAPIToken(ctx, input.TokenID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "pla_" + hex.EncodeToString(b)
}

// subscribeToFeed fetches and parses the feed URL to populate site URL and
// title, then creates the subscription. After creating the feed record, it
// processes the feed to persist entries immediately — so the feed is not
// empty after subscription. If the user already subscribes to the feed URL,
// the existing feed is returned (idempotent), making it safe for feed-list
// import.
func (a *API) subscribeToFeed(ctx context.Context, userID int, feedURL string, folderID *int) (*store.Feed, error) {
	if existing, err := a.store.GetFeedByURL(ctx, feedURL, userID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	siteURL := ""
	title := feedURL
	description := ""

	// Discover the actual feed URL. If the input is already a feed URL,
	// discovery returns it as-is. If the input is an HTML page, discovery
	// parses <link rel="alternate"> tags to find the feed URL.
	if a.fetcher != nil {
		if discovery, err := discoverer.Discover(ctx, a.fetcher, feedURL); err == nil {
			feedURL = discovery.FeedURL
			if discovery.SiteURL != "" {
				siteURL = discovery.SiteURL
			}
			// Check if we already subscribe to the discovered feed URL.
			if existing, err := a.store.GetFeedByURL(ctx, feedURL, userID); err == nil && existing != nil {
				return existing, nil
			}
			// Fetch and parse the feed to populate the real title and description.
			if result, err := a.fetcher.Fetch(ctx, feedURL, "", ""); err == nil && !result.NotModified {
				if parsed, err := parser.Parse(result.Body, result.ContentType); err == nil {
					if parsed.SiteURL != "" {
						siteURL = parsed.SiteURL
					}
					if parsed.Title != "" {
						title = parsed.Title
					}
					description = parsed.Description
				}
			}
		}
	}

	feed, err := a.store.CreateFeed(ctx, userID, folderID, feedURL, siteURL, title, description, store.SourceRSS)
	if err != nil {
		return nil, err
	}

	// Process the feed immediately so entries are persisted on subscribe.
	// Best-effort: the scheduler will retry on failure.
	if a.processor != nil {
		_ = a.processor.ProcessFeed(ctx, feed)
	}

	// Mark all entries as read so the user doesn't see a backlog of unread
	// items from before they subscribed.
	_ = a.store.MarkFeedEntriesRead(ctx, feed.ID, userID)

	return feed, nil
}

// --- Feed Lists ---

type CreateFeedListInput struct {
	Body struct {
		Title       string `json:"title" minLength:"1" maxLength:"255"`
		Description string `json:"description,omitempty" maxLength:"2000"`
		IsPublic    bool   `json:"is_public"`
	}
}

type UpdateFeedListInput struct {
	ListID int `path:"listId"`
	Body   struct {
		Title       string `json:"title" minLength:"1" maxLength:"255"`
		Description string `json:"description,omitempty" maxLength:"2000"`
		IsPublic    bool   `json:"is_public"`
	}
}

type AddFeedListFeedInput struct {
	ListID int `path:"listId"`
	Body   struct {
		FeedURL string `json:"feed_url" minLength:"1" maxLength:"2048"`
		SiteURL string `json:"site_url,omitempty" maxLength:"2048"`
		Title   string `json:"title,omitempty" maxLength:"512"`
	}
}

type FeedListDetailOutput struct {
	Body store.FeedList
}

type FeedListListOutput struct {
	Body []store.FeedList
}

func (a *API) registerFeedListRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-feed-list",
		Method:      http.MethodPost,
		Path:        "/api/v1/feed-lists",
		Summary:     "Create a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *CreateFeedListInput) (*FeedListDetailOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		fl, err := a.store.CreateFeedList(ctx, userID, input.Body.Title, input.Body.Description, input.Body.IsPublic)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &FeedListDetailOutput{Body: *fl}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-my-feed-lists",
		Method:      http.MethodGet,
		Path:        "/api/v1/feed-lists",
		Summary:     "List my feed lists",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, _ *struct{}) (*FeedListListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		lists, err := a.store.ListMyFeedLists(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if lists == nil {
			lists = []store.FeedList{}
		}
		return &FeedListListOutput{Body: lists}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-followed-feed-lists",
		Method:      http.MethodGet,
		Path:        "/api/v1/feed-lists/followed",
		Summary:     "List feed lists I follow",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, _ *struct{}) (*FeedListListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		lists, err := a.store.ListFollowedFeedLists(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if lists == nil {
			lists = []store.FeedList{}
		}
		return &FeedListListOutput{Body: lists}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "discover-feed-lists",
		Method:      http.MethodGet,
		Path:        "/api/v1/feed-lists/discover",
		Summary:     "Discover public feed lists",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		Limit  int `query:"limit" default:"24" maximum:"100"`
		Offset int `query:"offset" default:"0"`
	}) (*FeedListListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		if input.Limit == 0 {
			input.Limit = 24
		}
		lists, err := a.store.ListPublicFeedLists(ctx, userID, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if lists == nil {
			lists = []store.FeedList{}
		}
		return &FeedListListOutput{Body: lists}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "get-feed-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/feed-lists/{listId}",
		Summary:     "Get a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
	}) (*FeedListDetailOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		fl, err := a.store.GetFeedList(ctx, input.ListID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if fl == nil {
			return nil, huma.Error404NotFound("feed list not found")
		}
		feeds, err := a.store.ListFeedListFeeds(ctx, input.ListID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if feeds != nil {
			fl.Feeds = feeds
		} else {
			fl.Feeds = []store.FeedListFeed{}
		}
		return &FeedListDetailOutput{Body: *fl}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "update-feed-list",
		Method:      http.MethodPatch,
		Path:        "/api/v1/feed-lists/{listId}",
		Summary:     "Update a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *UpdateFeedListInput) (*FeedListDetailOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		fl, err := a.store.UpdateFeedList(ctx, input.ListID, userID, input.Body.Title, input.Body.Description, input.Body.IsPublic)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if fl == nil {
			return nil, huma.Error404NotFound("feed list not found")
		}
		return &FeedListDetailOutput{Body: *fl}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-feed-list",
		Method:      http.MethodDelete,
		Path:        "/api/v1/feed-lists/{listId}",
		Summary:     "Delete a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteFeedList(ctx, input.ListID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "add-feed-list-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feed-lists/{listId}/feeds",
		Summary:     "Add a feed to a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *AddFeedListFeedInput) (*struct {
		Body store.FeedListFeed
	}, error) {
		userID := auth.UserIDFromCtx(ctx)
		isOwner, err := a.store.IsFeedListOwner(ctx, input.ListID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if !isOwner {
			return nil, huma.Error403Forbidden("not the feed list owner")
		}
		flf, err := a.store.AddFeedListFeed(ctx, input.ListID, userID, input.Body.FeedURL, input.Body.SiteURL, input.Body.Title)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &struct {
			Body store.FeedListFeed
		}{Body: *flf}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "remove-feed-list-feed",
		Method:      http.MethodDelete,
		Path:        "/api/v1/feed-lists/{listId}/feeds/{itemId}",
		Summary:     "Remove a feed from a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
		ItemID int `path:"itemId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.RemoveFeedListFeed(ctx, input.ListID, input.ItemID, userID); err != nil {
			if errors.Is(err, store.ErrFeedListNotFound) {
				return nil, huma.Error404NotFound("feed list item not found")
			}
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "follow-feed-list",
		Method:      http.MethodPost,
		Path:        "/api/v1/feed-lists/{listId}/follow",
		Summary:     "Follow a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.FollowFeedList(ctx, input.ListID, userID); err != nil {
			switch {
			case errors.Is(err, store.ErrFeedListNotFound):
				return nil, huma.Error404NotFound("feed list not found")
			case errors.Is(err, store.ErrFeedListOwnList):
				return nil, huma.Error400BadRequest("cannot follow your own list")
			case errors.Is(err, store.ErrFeedListNotPublic):
				return nil, huma.Error403Forbidden("feed list is not public")
			}
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "unfollow-feed-list",
		Method:      http.MethodDelete,
		Path:        "/api/v1/feed-lists/{listId}/follow",
		Summary:     "Unfollow a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.UnfollowFeedList(ctx, input.ListID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "import-feed-list",
		Method:      http.MethodPost,
		Path:        "/api/v1/feed-lists/{listId}/import",
		Summary:     "Subscribe to all feeds in a feed list",
		Tags:        []string{"feed-lists"},
	}, func(ctx context.Context, input *struct {
		ListID int `path:"listId"`
	}) (*struct {
		Body struct {
			Imported int      `json:"imported"`
			Skipped  int      `json:"skipped"`
			Failed   int      `json:"failed"`
			FeedIDs  []int    `json:"feed_ids"`
			Errors   []string `json:"errors,omitempty"`
		}
	}, error) {
		userID := auth.UserIDFromCtx(ctx)
		fl, err := a.store.GetFeedList(ctx, input.ListID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if fl == nil {
			return nil, huma.Error404NotFound("feed list not found")
		}
		feeds, err := a.store.ListFeedListFeeds(ctx, input.ListID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		result := struct {
			Imported int      `json:"imported"`
			Skipped  int      `json:"skipped"`
			Failed   int      `json:"failed"`
			FeedIDs  []int    `json:"feed_ids"`
			Errors   []string `json:"errors,omitempty"`
		}{FeedIDs: []int{}}

		for _, flf := range feeds {
			// Idempotent: subscribeToFeed returns the existing feed if already subscribed.
			feed, fErr := a.subscribeToFeed(ctx, userID, flf.FeedURL, nil)
			if fErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, flf.FeedURL+": "+fErr.Error())
				continue
			}
			result.FeedIDs = append(result.FeedIDs, feed.ID)
			// Heuristic: if last_fetch_at is nil the feed is brand new.
			if feed.LastFetchAt == nil {
				result.Imported++
			} else {
				result.Skipped++
			}
		}
		return &struct {
			Body struct {
				Imported int      `json:"imported"`
				Skipped  int      `json:"skipped"`
				Failed   int      `json:"failed"`
				FeedIDs  []int    `json:"feed_ids"`
				Errors   []string `json:"errors,omitempty"`
			}
		}{Body: result}, nil
	})
}

// --- X (Twitter) Timelines ---

// SubscribeXFeedInput is the request body for subscribing to an X timeline.
// The user identifies the X account by @username; the server resolves it to
// the numeric user ID required by the X API v2.
type SubscribeXFeedInput struct {
	Body struct {
		Username string `json:"username" minLength:"1" maxLength:"255" doc:"X @username (with or without the leading @)"`
		FolderID *int   `json:"folder_id,omitempty"`
	}
}

// PreviewXFeedInput is the request body for previewing an X timeline.
type PreviewXFeedInput struct {
	Body struct {
		Username string `json:"username" minLength:"1" maxLength:"255" doc:"X @username (with or without the leading @)"`
	}
}

// PreviewXFeedBody is the preview response for an X timeline.
type PreviewXFeedBody struct {
	Title    string            `json:"title"`
	SiteURL  string            `json:"site_url"`
	FeedURL  string            `json:"feed_url"`
	Username string            `json:"username"`
	Items    []PreviewFeedItem `json:"items"`
}

// PreviewXFeedOutput wraps the X timeline preview response.
type PreviewXFeedOutput struct {
	Body PreviewXFeedBody
}

func (a *API) registerXRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "subscribe-x-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds/x",
		Summary:     "Subscribe to an X (Twitter) user timeline",
		Description: "Subscribes to a person's X timeline via the official X API v2. The username is resolved to a numeric user ID and a feed with source \"x\" is created. Entries are stored with entry_type \"post\" so consumers can render them differently from RSS articles.",
		Tags:        []string{"x"},
	}, func(ctx context.Context, input *SubscribeXFeedInput) (*FeedOutput, error) {
		if a.xclient == nil || !a.xclient.Enabled() {
			return nil, huma.Error503ServiceUnavailable("X API is not configured (set X_API_BEARER_TOKEN)")
		}
		userID := auth.UserIDFromCtx(ctx)
		feed, err := a.subscribeToXFeed(ctx, userID, input.Body.Username, input.Body.FolderID)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &FeedOutput{Body: *feed}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "preview-x-feed",
		Method:      http.MethodPost,
		Path:        "/api/v1/feeds/x/preview",
		Summary:     "Preview an X (Twitter) user timeline",
		Description: "Fetches a person's recent X posts via the official X API v2 without persisting anything.",
		Tags:        []string{"x"},
	}, func(ctx context.Context, input *PreviewXFeedInput) (*PreviewXFeedOutput, error) {
		if a.xclient == nil || !a.xclient.Enabled() {
			return nil, huma.Error503ServiceUnavailable("X API is not configured (set X_API_BEARER_TOKEN)")
		}
		user, err := a.xclient.ResolveUsername(ctx, input.Body.Username)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("could not resolve X user: %s", err.Error()))
		}
		parsed, err := a.xclient.FetchTimeline(ctx, user.ID)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("could not fetch X timeline: %s", err.Error()))
		}

		maxItems := 20
		if len(parsed.Items) > maxItems {
			parsed.Items = parsed.Items[:maxItems]
		}

		items := make([]PreviewFeedItem, 0, len(parsed.Items))
		for _, item := range parsed.Items {
			sanitized, err := sanitizer.Sanitize(item.Content)
			if err != nil {
				sanitized = item.Content
			}
			publishedAt, err := processor.ParseDate(item.PublishedAt)
			if err != nil {
				publishedAt = time.Now()
			}
			items = append(items, PreviewFeedItem{
				Title:       item.Title,
				URL:         item.Link,
				Author:      item.Author,
				Description: sanitizer.StripHTML(item.Description),
				Content:     sanitized,
				PublishedAt: publishedAt,
				Tags:        item.Tags,
			})
		}

		return &PreviewXFeedOutput{Body: PreviewXFeedBody{
			Title:    parsed.Title,
			SiteURL:  parsed.SiteURL,
			FeedURL:  user.ID,
			Username: user.Username,
			Items:    items,
		}}, nil
	})
}

// subscribeToXFeed resolves an X username to a numeric user ID, creates an
// X-source feed, and processes it immediately so posts are persisted on
// subscribe. The feed_url stores the numeric X user ID; the site_url stores
// the canonical x.com profile URL. If the user already subscribes to this X
// user, the existing feed is returned (idempotent).
func (a *API) subscribeToXFeed(ctx context.Context, userID int, username string, folderID *int) (*store.Feed, error) {
	user, err := a.xclient.ResolveUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Idempotent: if the user already subscribes to this X user ID, return it.
	if existing, err := a.store.GetFeedByURL(ctx, user.ID, userID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	siteURL := "https://x.com/" + user.Username
	title := user.Name
	if title == "" {
		title = "@" + user.Username
	}

	feed, err := a.store.CreateFeed(ctx, userID, folderID, user.ID, siteURL, title, user.Name, store.SourceX)
	if err != nil {
		return nil, err
	}

	// Process the feed immediately so entries are persisted on subscribe.
	// Best-effort: the scheduler will retry on failure.
	if a.processor != nil {
		_ = a.processor.ProcessFeed(ctx, feed)
	}

	// Mark all entries as read so the user doesn't see a backlog of unread
	// posts from before they subscribed.
	_ = a.store.MarkFeedEntriesRead(ctx, feed.ID, userID)

	return feed, nil
}

// --- OPML ---

// opmlXML is the XML representation of an OPML document.
type opmlXML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    opmlHead `xml:"head"`
	Body    opmlBody `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title,omitempty"`
}

type opmlBody struct {
	Outlines []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	XMLURL   string        `xml:"xmlUrl,attr,omitempty"`
	HTMLURL  string        `xml:"htmlUrl,attr,omitempty"`
	Title    string        `xml:"title,attr,omitempty"`
	Text     string        `xml:"text,attr,omitempty"`
	Type     string        `xml:"type,attr,omitempty"`
	Outlines []opmlOutline `xml:"outline,omitempty"`
}

type OPMLExportOutput struct {
	Body []byte
}

type OPMLImportInput struct {
	RawBody []byte
}

type OPMLImportResult struct {
	Body struct {
		Imported int      `json:"imported"`
		Skipped  int      `json:"skipped"`
		Failed   int      `json:"failed"`
		FeedIDs  []int    `json:"feed_ids"`
		Errors   []string `json:"errors,omitempty"`
	}
}

func (a *API) registerOPMLRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "export-opml",
		Method:      http.MethodGet,
		Path:        "/api/v1/opml/export",
		Summary:     "Export feeds as OPML",
		Description: "Returns all feed subscriptions and folders as an OPML XML document.",
		Tags:        []string{"opml"},
	}, func(ctx context.Context, _ *struct{}) (*OPMLExportOutput, error) {
		userID := auth.UserIDFromCtx(ctx)

		feeds, err := a.store.ListFeeds(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		folders, err := a.store.ListFolders(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		// Group feeds by folder. Feeds without a folder go into the root.
		folderMap := make(map[int][]store.Feed)
		var unfiled []store.Feed
		for _, f := range feeds {
			if f.FolderID != nil {
				folderMap[*f.FolderID] = append(folderMap[*f.FolderID], f)
			} else {
				unfiled = append(unfiled, f)
			}
		}

		var outlines []opmlOutline
		for _, fo := range folders {
			folderFeeds := folderMap[fo.ID]
			if len(folderFeeds) == 0 {
				continue
			}
			folderOutline := opmlOutline{
				Title:    fo.Title,
				Text:     fo.Title,
				Outlines: feedsToOutlines(folderFeeds),
			}
			outlines = append(outlines, folderOutline)
		}
		outlines = append(outlines, feedsToOutlines(unfiled)...)

		doc := opmlXML{
			Version: "2.0",
			Head:    opmlHead{Title: "Earthed Subscriptions"},
			Body:    opmlBody{Outlines: outlines},
		}

		data, err := xml.MarshalIndent(doc, "", "  ")
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		data = append([]byte(xml.Header), data...)

		return &OPMLExportOutput{Body: data}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "import-opml",
		Method:      http.MethodPost,
		Path:        "/api/v1/opml/import",
		Summary:     "Import feeds from an OPML file",
		Description: "Parses an OPML XML document and subscribes the user to all feeds found. Folders are created as needed. Existing subscriptions are skipped.",
		Tags:        []string{"opml"},
	}, func(ctx context.Context, input *OPMLImportInput) (*OPMLImportResult, error) {
		userID := auth.UserIDFromCtx(ctx)

		var doc opmlXML
		decoder := xml.NewDecoder(bytes.NewReader(input.RawBody))
		decoder.Strict = false
		if err := decoder.Decode(&doc); err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("invalid OPML: %s", err.Error()))
		}

		// Build folder name → ID map from existing folders.
		existingFolders, err := a.store.ListFolders(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		folderByName := make(map[string]int)
		for _, fo := range existingFolders {
			folderByName[fo.Title] = fo.ID
		}

		result := OPMLImportResult{}
		result.Body.FeedIDs = []int{}

		getOrCreateFolder := func(name string) (*int, error) {
			if name == "" {
				return nil, nil
			}
			if id, ok := folderByName[name]; ok {
				return &id, nil
			}
			folder, err := a.store.CreateFolder(ctx, userID, name, nil)
			if err != nil {
				return nil, err
			}
			folderByName[name] = folder.ID
			return &folder.ID, nil
		}

		var processOutlines func(outlines []opmlOutline, folderName string)
		processOutlines = func(outlines []opmlOutline, folderName string) {
			for _, o := range outlines {
				if o.XMLURL != "" {
					// Leaf node — a feed subscription.
					var folderID *int
					if folderName != "" {
						f, err := getOrCreateFolder(folderName)
						if err != nil {
							result.Body.Failed++
							result.Body.Errors = append(result.Body.Errors, o.XMLURL+": "+err.Error())
							continue
						}
						folderID = f
					}

					feed, fErr := a.subscribeToFeed(ctx, userID, o.XMLURL, folderID)
					if fErr != nil {
						result.Body.Failed++
						result.Body.Errors = append(result.Body.Errors, o.XMLURL+": "+fErr.Error())
						continue
					}
					result.Body.FeedIDs = append(result.Body.FeedIDs, feed.ID)
					if feed.LastFetchAt == nil {
						result.Body.Imported++
					} else {
						result.Body.Skipped++
					}
				} else {
					// Container node — a folder. Use its title, falling back to text.
					name := o.Title
					if name == "" {
						name = o.Text
					}
					processOutlines(o.Outlines, name)
				}
			}
		}

		processOutlines(doc.Body.Outlines, "")

		return &result, nil
	})
}

// feedsToOutlines converts a slice of feeds to OPML outline elements.
func feedsToOutlines(feeds []store.Feed) []opmlOutline {
	outlines := make([]opmlOutline, 0, len(feeds))
	for _, f := range feeds {
		outlines = append(outlines, opmlOutline{
			XMLURL:  f.FeedURL,
			HTMLURL: f.SiteURL,
			Title:   f.Title,
			Text:    f.Title,
			Type:    "rss",
		})
	}
	return outlines
}
