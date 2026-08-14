// Package api registers the Planetary REST API on a huma router.
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/fuegoio/planetary/internal/auth"
	"github.com/fuegoio/planetary/internal/store"
)

// API wires the store and auth to a huma router and registers the REST routes.
type API struct {
	huma   huma.API
	store  *store.Store
	auth   *auth.Auth
}

// New returns an API bound to the given huma router, store, and auth.
func New(humaAPI huma.API, st *store.Store, authInst *auth.Auth) *API {
	return &API{huma: humaAPI, store: st, auth: authInst}
}

// OpenAPITags returns the ordered tag list for the OpenAPI spec.
func OpenAPITags() []*huma.Tag {
	return []*huma.Tag{
		{Name: "feeds", Description: "Feed subscriptions"},
		{Name: "entries", Description: "Feed entries/articles"},
		{Name: "categories", Description: "Feed categories"},
		{Name: "users", Description: "User accounts"},
		{Name: "tokens", Description: "API tokens"},
		{Name: "opml", Description: "OPML import/export"},
	}
}

// RegisterRoutes registers all Planetary REST routes on the huma router.
func (a *API) RegisterRoutes() {
	a.registerHealthRoutes()
	a.registerMeRoutes()
	a.registerCategoryRoutes()
	a.registerFeedRoutes()
	a.registerEntryRoutes()
	a.registerTokenRoutes()
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
}

// --- Categories ---

type CreateCategoryInput struct {
	Body struct {
		Title string `json:"title" minLength:"1" maxLength:"255"`
	}
}

type CategoryOutput struct {
	Body store.Category
}

type CategoryListOutput struct {
	Body []store.Category
}

func (a *API) registerCategoryRoutes() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-category",
		Method:      http.MethodPost,
		Path:        "/api/v1/categories",
		Summary:     "Create a category",
		Tags:        []string{"categories"},
	}, func(ctx context.Context, input *CreateCategoryInput) (*CategoryOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		cat, err := a.store.CreateCategory(ctx, userID, input.Body.Title)
		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return &CategoryOutput{Body: *cat}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "list-categories",
		Method:      http.MethodGet,
		Path:        "/api/v1/categories",
		Summary:     "List categories",
		Tags:        []string{"categories"},
	}, func(ctx context.Context, _ *struct{}) (*CategoryListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		cats, err := a.store.ListCategories(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		if cats == nil {
			cats = []store.Category{}
		}
		return &CategoryListOutput{Body: cats}, nil
	})

	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-category",
		Method:      http.MethodDelete,
		Path:        "/api/v1/categories/{categoryId}",
		Summary:     "Delete a category",
		Tags:        []string{"categories"},
	}, func(ctx context.Context, input *struct {
		CategoryID int `path:"categoryId"`
	}) (*struct{}, error) {
		userID := auth.UserIDFromCtx(ctx)
		if err := a.store.DeleteCategory(ctx, input.CategoryID, userID); err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		return nil, nil
	})
}

// --- Feeds ---

type CreateFeedInput struct {
	Body struct {
		FeedURL    string `json:"feed_url" minLength:"1" maxLength:"2048"`
		CategoryID *int   `json:"category_id,omitempty"`
	}
}

type FeedOutput struct {
	Body store.Feed
}

type FeedListOutput struct {
	Body []store.Feed
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
		feed, err := a.store.CreateFeed(ctx, userID, input.Body.CategoryID, input.Body.FeedURL, "", input.Body.FeedURL)
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
		FeedID     *int   `query:"feed_id"`
		CategoryID *int   `query:"category_id"`
		Status     string `query:"status" enum:"unread,read,removed"`
		Starred    *bool  `query:"starred"`
		Search     string `query:"search"`
		Limit      int    `query:"limit" default:"50" maximum:"200"`
		Offset     int    `query:"offset" default:"0"`
	}) (*EntryListOutput, error) {
		userID := auth.UserIDFromCtx(ctx)
		if input.Limit == 0 {
			input.Limit = 50
		}
		entries, err := a.store.ListEntries(ctx, userID, input.FeedID, input.CategoryID, input.Status, input.Starred, input.Search, input.Limit, input.Offset)
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
		t, err := a.store.CreateAPIToken(ctx, userID, input.Body.Label, hash)
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

// uuid import used by future OPML import/export
var _ = uuid.New
