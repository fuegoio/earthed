// Package fanout manages PDS WebSocket subscriptions and aggregates io.planetary.* events.
package fanout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/fuegoio/planetary/go/relay/internal/store"
)

// Fanout manages one goroutine per tracked DID that subscribes to its PDS repo stream.
type Fanout struct {
	store          *store.Store
	reconnectDelay time.Duration
	httpClient     *http.Client
	subscribers    map[string]chan *store.RelayEvent
	subsMu         sync.RWMutex
	workers        map[string]context.CancelFunc
	wmu            sync.Mutex
}

// New returns a Fanout.
func New(st *store.Store, reconnectDelay time.Duration) *Fanout {
	return &Fanout{
		store:          st,
		reconnectDelay: reconnectDelay,
		httpClient:     &http.Client{Timeout: 15 * time.Second},
		subscribers:    make(map[string]chan *store.RelayEvent),
		workers:        make(map[string]context.CancelFunc),
	}
}

// Start loads all active tracked DIDs and begins subscribing. Blocks until ctx is done.
func (f *Fanout) Start(ctx context.Context) {
	dids, err := f.store.ListActiveTrackedDIDs(ctx)
	if err != nil {
		slog.Error("fanout: list active dids", "err", err)
		return
	}
	for _, d := range dids {
		f.EnsureSubscribed(ctx, d.DID, d.PDSUrl, d.CursorSeq)
	}
	<-ctx.Done()
}

// EnsureSubscribed starts a subscription goroutine for did if none is running.
func (f *Fanout) EnsureSubscribed(parentCtx context.Context, did, pdsURL string, cursorSeq int64) {
	f.wmu.Lock()
	defer f.wmu.Unlock()
	if _, ok := f.workers[did]; ok {
		return
	}
	workerCtx, cancel := context.WithCancel(parentCtx)
	f.workers[did] = cancel
	go f.runWorker(workerCtx, did, pdsURL, cursorSeq)
}

// Subscribe registers an instance listener and returns its event channel.
func (f *Fanout) Subscribe(instanceURL string) chan *store.RelayEvent {
	ch := make(chan *store.RelayEvent, 256)
	f.subsMu.Lock()
	f.subscribers[instanceURL] = ch
	f.subsMu.Unlock()
	return ch
}

// Unsubscribe removes an instance listener.
func (f *Fanout) Unsubscribe(instanceURL string) {
	f.subsMu.Lock()
	delete(f.subscribers, instanceURL)
	f.subsMu.Unlock()
}

func (f *Fanout) runWorker(ctx context.Context, did, pdsURL string, cursorSeq int64) {
	slog.Info("fanout: starting worker", "did", did)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := f.subscribe(ctx, did, pdsURL, &cursorSeq); err != nil && ctx.Err() == nil {
			slog.Warn("fanout: subscription error", "did", did, "err", err)
			_ = f.store.SetTrackedDIDError(context.Background(), did, err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(f.reconnectDelay):
		}
	}
}

// reposFrame is the JSON shape of a com.atproto.sync.subscribeRepos #commit frame.
// AT Proto PDSes emit CBOR; we request JSON via Accept header for simplicity.
// Production deployments should add CBOR decoding.
type reposFrame struct {
	Type string   `json:"$type"`
	Seq  int64    `json:"seq"`
	Repo string   `json:"repo"`
	Ops  []repoOp `json:"ops"`
}

type repoOp struct {
	Action string `json:"action"` // create | update | delete
	Path   string `json:"path"`   // <collection>/<rkey>
}

func (f *Fanout) subscribe(ctx context.Context, did, pdsURL string, cursorSeq *int64) error {
	u, err := url.Parse(pdsURL)
	if err != nil {
		return fmt.Errorf("parse pds url: %w", err)
	}
	wsScheme := "wss"
	if u.Scheme == "http" {
		wsScheme = "ws"
	}
	wsURL := fmt.Sprintf("%s://%s/xrpc/com.atproto.sync.subscribeRepos?wantedDids=%s",
		wsScheme, u.Host, url.QueryEscape(did))
	if *cursorSeq > 0 {
		wsURL += fmt.Sprintf("&cursor=%d", *cursorSeq)
	}
	ws, err := websocket.Dial(wsURL, "", pdsURL)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = ws.Close() }()
	slog.Info("fanout: connected to PDS", "did", did)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		var raw json.RawMessage
		if err := websocket.JSON.Receive(ws, &raw); err != nil {
			return fmt.Errorf("receive: %w", err)
		}
		var frame reposFrame
		if err := json.Unmarshal(raw, &frame); err != nil {
			continue
		}
		if !strings.HasSuffix(frame.Type, "#commit") {
			continue
		}
		for _, op := range frame.Ops {
			f.processOp(context.Background(), frame.Repo, pdsURL, op)
		}
		if frame.Seq > 0 {
			_ = f.store.UpdateTrackedDIDCursor(context.Background(), did, frame.Seq)
			*cursorSeq = frame.Seq
		}
	}
}

func (f *Fanout) processOp(ctx context.Context, did, pdsURL string, op repoOp) {
	parts := strings.SplitN(op.Path, "/", 2)
	if len(parts) != 2 {
		return
	}
	col, rkey := parts[0], parts[1]
	switch col {
	case "io.planetary.graph.follow":
		f.handleFollow(ctx, did, pdsURL, rkey, op.Action)
	case "io.planetary.share.article":
		f.handleShare(ctx, did, pdsURL, rkey, op.Action)
	case "io.planetary.feed.subscription":
		f.handleFeedSub(ctx, did, pdsURL, rkey, op.Action)
	}
}

func (f *Fanout) handleFollow(ctx context.Context, did, pdsURL, rkey, action string) {
	if action == "delete" {
		followeeDID, deleted, err := f.store.DeleteFollow(ctx, did, rkey)
		if err != nil {
			slog.Warn("fanout: delete follow", "err", err)
			return
		}
		if deleted {
			f.emit(ctx, "unfollow", did, map[string]any{"did": did, "subjectDid": followeeDID, "rkey": rkey})
		}
		return
	}
	rec, err := f.fetchRecord(ctx, pdsURL, did, "io.planetary.graph.follow", rkey)
	if err != nil {
		slog.Warn("fanout: fetch follow", "err", err)
		return
	}
	subject, _ := rec["subject"].(string)
	if subject == "" {
		return
	}
	createdAt := parseTime(rec["createdAt"])
	isNew, err := f.store.RecordFollow(ctx, did, subject, rkey, pdsURL, createdAt)
	if err != nil {
		slog.Warn("fanout: record follow", "err", err)
		return
	}
	if isNew {
		f.emit(ctx, "follow", did, map[string]any{
			"did": did, "subjectDid": subject, "rkey": rkey, "createdAt": createdAt,
		})
	}
}

func (f *Fanout) handleShare(ctx context.Context, did, pdsURL, rkey, action string) {
	if action == "delete" {
		deleted, err := f.store.DeleteShare(ctx, did, rkey)
		if err != nil {
			slog.Warn("fanout: delete share", "err", err)
			return
		}
		if deleted {
			f.emit(ctx, "unshare", did, map[string]any{"did": did, "rkey": rkey})
		}
		return
	}
	rec, err := f.fetchRecord(ctx, pdsURL, did, "io.planetary.share.article", rkey)
	if err != nil {
		slog.Warn("fanout: fetch share", "err", err)
		return
	}
	articleURL, _ := rec["articleUrl"].(string)
	feedURL, _ := rec["feedUrl"].(string)
	title, _ := rec["title"].(string)
	sharedAt := parseTime(rec["sharedAt"])
	isNew, err := f.store.RecordShare(ctx, did, rkey, articleURL, feedURL, title, pdsURL, &sharedAt)
	if err != nil {
		slog.Warn("fanout: record share", "err", err)
		return
	}
	if isNew {
		f.emit(ctx, "share", did, map[string]any{
			"did": did, "rkey": rkey, "articleUrl": articleURL,
			"feedUrl": feedURL, "title": title, "sharedAt": sharedAt,
		})
	}
}

func (f *Fanout) handleFeedSub(ctx context.Context, did, pdsURL, rkey, action string) {
	if action == "delete" {
		deleted, err := f.store.DeleteFeedSubscription(ctx, did, rkey)
		if err != nil {
			slog.Warn("fanout: delete feed sub", "err", err)
			return
		}
		if deleted {
			f.emit(ctx, "feedUnsubscription", did, map[string]any{"did": did, "rkey": rkey})
		}
		return
	}
	rec, err := f.fetchRecord(ctx, pdsURL, did, "io.planetary.feed.subscription", rkey)
	if err != nil {
		slog.Warn("fanout: fetch feed sub", "err", err)
		return
	}
	feedURL, _ := rec["feedUrl"].(string)
	createdAt := parseTime(rec["createdAt"])
	isNew, err := f.store.RecordFeedSubscription(ctx, did, rkey, feedURL, pdsURL, &createdAt)
	if err != nil {
		slog.Warn("fanout: record feed sub", "err", err)
		return
	}
	if isNew {
		f.emit(ctx, "feedSubscription", did, map[string]any{
			"did": did, "rkey": rkey, "feedUrl": feedURL, "createdAt": createdAt,
		})
	}
}

func (f *Fanout) emit(ctx context.Context, eventType, did string, payload any) {
	seq, err := f.store.AppendEvent(ctx, eventType, did, payload)
	if err != nil {
		slog.Warn("fanout: append event", "err", err)
		return
	}
	b, _ := json.Marshal(payload)
	evt := &store.RelayEvent{Seq: seq, EventType: eventType, DID: did, Payload: b}
	f.subsMu.RLock()
	defer f.subsMu.RUnlock()
	for _, ch := range f.subscribers {
		select {
		case ch <- evt:
		default:
		}
	}
}

type getRecordResp struct {
	Value map[string]any `json:"value"`
}

func (f *Fanout) fetchRecord(ctx context.Context, pdsURL, did, collection, rkey string) (map[string]any, error) {
	u := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pdsURL, url.QueryEscape(did), url.QueryEscape(collection), url.QueryEscape(rkey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pds returned %d for %s/%s", resp.StatusCode, collection, rkey)
	}
	var out getRecordResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

func parseTime(v any) time.Time {
	s, _ := v.(string)
	if s == "" {
		return time.Now().UTC()
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}
