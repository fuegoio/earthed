package atproto

import (
	"context"
	"fmt"
	"time"
)

// Writer performs high-level AT Proto record writes for a single user's repo.
// It wraps a Client with the user's DID and a token-refresh callback.
type Writer struct {
	client *Client
	did    string
}

// NewWriter returns a Writer for the given DID and PDS.
func NewWriter(pdsURL, did, accessToken string) *Writer {
	return &Writer{
		client: NewClient(pdsURL, accessToken),
		did:    did,
	}
}

// PutFollow writes an io.sunred.graph.follow record and returns the rkey.
func (w *Writer) PutFollow(ctx context.Context, subjectDID string) (string, error) {
	rkey := NewTID()
	_, err := w.client.PutRecord(ctx, w.did, CollectionFollow, rkey, FollowRecord{
		Type:      CollectionFollow,
		Subject:   subjectDID,
		CreatedAt: FormatTime(time.Now()),
	})
	if err != nil {
		return "", fmt.Errorf("put follow: %w", err)
	}
	return rkey, nil
}

// DeleteFollow removes the io.sunred.graph.follow record at rkey.
func (w *Writer) DeleteFollow(ctx context.Context, rkey string) error {
	return w.client.DeleteRecord(ctx, w.did, CollectionFollow, rkey)
}

// PutShare writes an io.sunred.share.article record and returns the rkey.
func (w *Writer) PutShare(ctx context.Context,
	articleURL, title, description,
	feedURL, feedTitle, feedSiteURL, author string,
	publishedAt *time.Time, sharedAt time.Time,
) (string, error) {
	rkey := NewTID()
	rec := ShareRecord{
		Type:        CollectionShare,
		ArticleURL:  articleURL,
		Title:       title,
		Description: description,
		FeedURL:     feedURL,
		FeedTitle:   feedTitle,
		FeedSiteURL: feedSiteURL,
		Author:      author,
		SharedAt:    FormatTime(sharedAt),
	}
	if publishedAt != nil {
		rec.PublishedAt = FormatTime(*publishedAt)
	}
	_, err := w.client.PutRecord(ctx, w.did, CollectionShare, rkey, rec)
	if err != nil {
		return "", fmt.Errorf("put share: %w", err)
	}
	return rkey, nil
}

// DeleteShare removes the io.sunred.share.article record at rkey.
func (w *Writer) DeleteShare(ctx context.Context, rkey string) error {
	return w.client.DeleteRecord(ctx, w.did, CollectionShare, rkey)
}

// PutFeedSubscription writes an io.sunred.feed.subscription record
// and returns the rkey.
func (w *Writer) PutFeedSubscription(ctx context.Context, feedURL, siteURL, title string, createdAt time.Time) (string, error) {
	rkey := NewTID()
	_, err := w.client.PutRecord(ctx, w.did, CollectionSubscription, rkey, SubscriptionRecord{
		Type:      CollectionSubscription,
		FeedURL:   feedURL,
		SiteURL:   siteURL,
		Title:     title,
		CreatedAt: FormatTime(createdAt),
	})
	if err != nil {
		return "", fmt.Errorf("put feed subscription: %w", err)
	}
	return rkey, nil
}

// DeleteFeedSubscription removes the io.sunred.feed.subscription record.
func (w *Writer) DeleteFeedSubscription(ctx context.Context, rkey string) error {
	return w.client.DeleteRecord(ctx, w.did, CollectionSubscription, rkey)
}

// PutFeedList writes or replaces an io.sunred.feed.list record.
// If rkey is empty, a new TID is generated.
func (w *Writer) PutFeedList(ctx context.Context, rkey, title, description string, isPublic bool, feeds []FeedListEntry, createdAt time.Time) (string, error) {
	if rkey == "" {
		rkey = NewTID()
	}
	_, err := w.client.PutRecord(ctx, w.did, CollectionFeedList, rkey, FeedListRecord{
		Type:        CollectionFeedList,
		Title:       title,
		Description: description,
		IsPublic:    isPublic,
		Feeds:       feeds,
		CreatedAt:   FormatTime(createdAt),
	})
	if err != nil {
		return "", fmt.Errorf("put feed list: %w", err)
	}
	return rkey, nil
}

// DeleteFeedList removes the io.sunred.feed.list record.
func (w *Writer) DeleteFeedList(ctx context.Context, rkey string) error {
	return w.client.DeleteRecord(ctx, w.did, CollectionFeedList, rkey)
}
