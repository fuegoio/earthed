// Package fetcher handles HTTP fetching of feed URLs with conditional
// requests (ETag / Last-Modified) and configurable timeouts.
package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Result holds the response of a feed fetch.
type Result struct {
	StatusCode    int
	Body          []byte
	ETag          string
	LastModified  string
	ContentType   string
	NotModified   bool
}

// Fetcher fetches feed URLs with configurable timeout and conditional headers.
type Fetcher struct {
	client    *http.Client
	maxBody   int64
	userAgent string
}

// New returns a Fetcher with the given timeout and max body size.
func New(timeout time.Duration, maxBody int64, userAgent string) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		maxBody:   maxBody,
		userAgent: userAgent,
	}
}

// Fetch performs a conditional GET on feedURL. If etag or lastModified are
// non-empty, they are sent as If-None-Match / If-Modified-Since respectively.
// A 304 response returns NotModified=true with no body.
func (f *Fetcher) Fetch(ctx context.Context, feedURL, etag, lastModified string) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if f.userAgent != "" {
		req.Header.Set("User-Agent", f.userAgent)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, application/json, */*")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", feedURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		return &Result{StatusCode: 304, NotModified: true}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", feedURL, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.maxBody))
	if err != nil {
		return nil, fmt.Errorf("read body %s: %w", feedURL, err)
	}

	return &Result{
		StatusCode:   resp.StatusCode,
		Body:         body,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		ContentType:  resp.Header.Get("Content-Type"),
	}, nil
}
