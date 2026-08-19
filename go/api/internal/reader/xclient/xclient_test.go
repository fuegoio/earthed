package xclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at a test server, with a dummy
// bearer token so Enabled() is true.
func newTestClient(t *testing.T, fn http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(fn)
	t.Cleanup(srv.Close)
	c := New(srv.URL, "test-bearer", "Planetary-test/1.0", 5*time.Second)
	return c, srv
}

func TestEnabled(t *testing.T) {
	c := New("", "", "", time.Second)
	if c.Enabled() {
		t.Fatal("client with empty token should not be enabled")
	}
	c = New("", "tok", "", time.Second)
	if !c.Enabled() {
		t.Fatal("client with token should be enabled")
	}
}

func TestFetchTimelineNotConfigured(t *testing.T) {
	c := New("", "", "", time.Second)
	if _, err := c.FetchTimeline(context.Background(), "123"); err != ErrNotConfigured {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}

func TestResolveUsername(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-bearer" {
			t.Errorf("Authorization = %q, want Bearer test-bearer", got)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/by/username/elonmusk") {
			t.Errorf("path = %q, want suffix /users/by/username/elonmusk", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(userResponse{
			Data: &struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				Username        string `json:"username"`
				ProfileImageURL string `json:"profile_image_url"`
			}{ID: "44196397", Name: "Elon Musk", Username: "elonmusk"},
		})
	})

	user, err := c.ResolveUsername(context.Background(), "@elonmusk")
	if err != nil {
		t.Fatalf("ResolveUsername: %v", err)
	}
	if user.ID != "44196397" {
		t.Errorf("ID = %q, want 44196397", user.ID)
	}
	if user.Username != "elonmusk" {
		t.Errorf("Username = %q, want elonmusk", user.Username)
	}
}

func TestResolveUsernameNotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{
				{"title": "Not Found", "detail": "User not found.", "status": 404},
			},
		})
	})
	if _, err := c.ResolveUsername(context.Background(), "nope"); err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestFetchTimeline(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "max_results=100") {
			t.Errorf("query missing max_results=100: %q", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "exclude=replies") {
			t.Errorf("query missing exclude=replies: %q", r.URL.RawQuery)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/44196397/tweets") {
			t.Errorf("path = %q, want suffix /users/44196397/tweets", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":         "1700000000000000000",
					"text":       "Hello world https://t.co/abc",
					"author_id":  "44196397",
					"created_at": "2024-01-01T12:00:00.000Z",
					"entities": map[string]any{
						"urls": []map[string]any{
							{
								"url":          "https://t.co/abc",
								"expanded_url": "https://example.com/post",
							},
						},
						"hashtags": []map[string]any{
							{"tag": "SpaceX"},
							{"tag": "Mars"},
						},
					},
				},
			},
			"includes": map[string]any{
				"users": []map[string]any{
					{
						"id":       "44196397",
						"name":     "Elon Musk",
						"username": "elonmusk",
					},
				},
			},
			"meta": map[string]any{"result_count": 1},
		})
	})

	feed, err := c.FetchTimeline(context.Background(), "44196397")
	if err != nil {
		t.Fatalf("FetchTimeline: %v", err)
	}
	if feed.Title != "Elon Musk" {
		t.Errorf("Title = %q, want Elon Musk", feed.Title)
	}
	if feed.SiteURL != "https://x.com/elonmusk" {
		t.Errorf("SiteURL = %q, want https://x.com/elonmusk", feed.SiteURL)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("Items = %d, want 1", len(feed.Items))
	}
	item := feed.Items[0]
	if item.Link != "https://x.com/i/status/1700000000000000000" {
		t.Errorf("Link = %q", item.Link)
	}
	// t.co link should be expanded.
	if !strings.Contains(item.Content, "https://example.com/post") {
		t.Errorf("Content = %q, want expanded URL", item.Content)
	}
	if strings.Contains(item.Content, "t.co") {
		t.Errorf("Content = %q, t.co link not expanded", item.Content)
	}
	if item.Author != "Elon Musk" {
		t.Errorf("Author = %q, want Elon Musk", item.Author)
	}
	if len(item.Tags) != 2 || item.Tags[0] != "SpaceX" || item.Tags[1] != "Mars" {
		t.Errorf("Tags = %v, want [SpaceX Mars]", item.Tags)
	}
	wantTime := "2024-01-01T12:00:00.000Z"
	if item.PublishedAt != wantTime {
		t.Errorf("PublishedAt = %q, want %q", item.PublishedAt, wantTime)
	}
}

func TestFetchTimelineEmpty(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
			"meta": map[string]any{"result_count": 0},
		})
	})
	feed, err := c.FetchTimeline(context.Background(), "123")
	if err != nil {
		t.Fatalf("FetchTimeline: %v", err)
	}
	if len(feed.Items) != 0 {
		t.Errorf("Items = %d, want 0", len(feed.Items))
	}
}

func TestFetchTimelineAPIError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":  "Unauthorized",
			"detail": "Invalid bearer token",
		})
	})
	if _, err := c.FetchTimeline(context.Background(), "123"); err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func mustDecodeEntities(t *testing.T, raw string) *entities {
	t.Helper()
	var ents entities
	if err := json.Unmarshal([]byte(raw), &ents); err != nil {
		t.Fatalf("unmarshal entities: %v", err)
	}
	return &ents
}

func TestExpandURLs(t *testing.T) {
	ents := mustDecodeEntities(t, `{
	  "urls": [{"url": "https://t.co/abc", "expanded_url": "https://example.com/expanded", "unwound_url": "https://example.com/unwound"}]
	}`)
	got := expandURLs("check this https://t.co/abc out", ents)
	if !strings.Contains(got, "https://example.com/unwound") {
		t.Errorf("expected unwound URL, got %q", got)
	}
}

func TestExpandURLsNilEntities(t *testing.T) {
	got := expandURLs("no links here", nil)
	if got != "no links here" {
		t.Errorf("got %q, want unchanged", got)
	}
}

func TestCollectTags(t *testing.T) {
	ents := mustDecodeEntities(t, `{"hashtags": [{"tag": "Go"}, {"tag": "X"}, {"tag": ""}]}`)
	tags := collectTags(ents)
	if len(tags) != 2 || tags[0] != "Go" || tags[1] != "X" {
		t.Errorf("tags = %v, want [Go X]", tags)
	}
}

func TestCollectTagsNil(t *testing.T) {
	if tags := collectTags(nil); tags != nil {
		t.Errorf("tags = %v, want nil", tags)
	}
}

func runeLen(s string) int { return len([]rune(s)) }

func TestTruncateForTitle(t *testing.T) {
	short := "a short post"
	if got := truncateForTitle(short); got != short {
		t.Errorf("got %q, want unchanged short text", got)
	}
	long := strings.Repeat("a", 200)
	got := truncateForTitle(long)
	if runeLen(got) > 140 {
		t.Errorf("title rune length = %d, want <= 140", runeLen(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("title should end with ellipsis, got %q", got)
	}
	// newlines should be collapsed to spaces.
	multiline := "line one\nline two"
	if got := truncateForTitle(multiline); strings.Contains(got, "\n") {
		t.Errorf("title should not contain newlines, got %q", got)
	}
}
