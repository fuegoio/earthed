// Package server implements the Sunred relay HTTP server.
//
// Endpoints:
//
//	POST  /xrpc/io.sunred.relay.announceUser   — instance registers a new DID
//	GET   /xrpc/io.sunred.relay.getCounts      — global counts for a DID
//	GET   /xrpc/io.sunred.relay.subscribeEvents — WebSocket event stream for instances
//	GET   /health
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/websocket"

	"github.com/fuegoio/sunred/go/relay/internal/fanout"
	"github.com/fuegoio/sunred/go/relay/internal/store"
)

// Server is the relay HTTP server.
type Server struct {
	store  *store.Store
	fanout *fanout.Fanout
}

// New returns a Server.
func New(st *store.Store, f *fanout.Fanout) *Server {
	return &Server{store: st, fanout: f}
}

// Handler returns the http.Handler for the relay server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/xrpc/io.sunred.relay.announceUser", s.handleAnnounceUser)
	mux.HandleFunc("/xrpc/io.sunred.relay.getCounts", s.handleGetCounts)
	mux.HandleFunc("/xrpc/io.sunred.relay.searchDIDs", s.handleSearchDIDs)
	mux.HandleFunc("/xrpc/io.sunred.relay.resolveHandle", s.handleResolveHandle)
	mux.Handle("/xrpc/io.sunred.relay.subscribeEvents", websocket.Handler(s.handleSubscribeEvents))
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// --- announceUser ---

type announceUserInput struct {
	DID         string `json:"did"`
	PDSUrl      string `json:"pdsUrl"`
	InstanceURL string `json:"instanceUrl"`
	Handle      string `json:"handle"`
}

type announceUserOutput struct {
	Tracked bool `json:"tracked"`
}

func (s *Server) handleAnnounceUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in announceUserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if in.DID == "" || in.PDSUrl == "" || in.InstanceURL == "" {
		http.Error(w, "did, pdsUrl, and instanceUrl are required", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	instanceID, err := s.store.UpsertInstance(ctx, in.InstanceURL)
	if err != nil {
		slog.Error("relay: upsert instance", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, isNew, err := s.store.UpsertTrackedDID(ctx, in.DID, in.PDSUrl, in.Handle, instanceID)
	if err != nil {
		slog.Error("relay: upsert tracked did", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if isNew {
		slog.Info("relay: tracking new DID", "did", in.DID, "pds", in.PDSUrl)
		// Backfill existing records from the PDS, then start the live
		// firehose subscription (tap-style backfill-then-cutover).
		go s.fanout.BackfillAndSubscribe(context.Background(), in.DID, in.PDSUrl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(announceUserOutput{Tracked: true})
}

// --- getCounts ---

type getCountsOutput struct {
	DID                   string `json:"did"`
	FollowerCount         int64  `json:"followerCount"`
	ShareCount            int64  `json:"shareCount"`
	FeedSubscriptionCount int64  `json:"feedSubscriptionCount"`
}

func (s *Server) handleGetCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	did := r.URL.Query().Get("did")
	if did == "" {
		http.Error(w, "did is required", http.StatusBadRequest)
		return
	}

	followers, shares, feedSubs, err := s.store.GetCounts(r.Context(), did)
	if err != nil {
		slog.Error("relay: get counts", "did", did, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(getCountsOutput{
		DID:                   did,
		FollowerCount:         followers,
		ShareCount:            shares,
		FeedSubscriptionCount: feedSubs,
	})
}

// --- searchDIDs ---

type searchDIDsOutput struct {
	Results []store.SearchResult `json:"results"`
}

func (s *Server) handleSearchDIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q is required", http.StatusBadRequest)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	results, err := s.store.SearchDIDs(r.Context(), q, limit)
	if err != nil {
		slog.Error("relay: search dids", "q", q, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []store.SearchResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(searchDIDsOutput{Results: results})
}

// --- resolveHandle ---

type resolveHandleOutput struct {
	DID    string `json:"did"`
	PDSUrl string `json:"pdsUrl"`
}

func (s *Server) handleResolveHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "handle is required", http.StatusBadRequest)
		return
	}
	did, pdsURL, err := s.store.ResolveHandle(r.Context(), handle)
	if err != nil {
		slog.Error("relay: resolve handle", "handle", handle, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if did == "" {
		http.Error(w, "handle not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resolveHandleOutput{DID: did, PDSUrl: pdsURL})
}

// --- subscribeEvents ---
// to the subscribing instance. Supports cursor-based replay.
//
// To avoid the replay/subscribe race, the instance is registered for live
// events BEFORE replaying. Events that arrive during replay are held in the
// subscriber channel's buffer (capacity 256). After replay completes, the
// main read loop drains the buffer and continues with live events. Any
// events already replayed are skipped by seq comparison.
func (s *Server) handleSubscribeEvents(ws *websocket.Conn) {
	r := ws.Request()
	instanceURL := r.URL.Query().Get("instanceUrl")
	cursorStr := r.URL.Query().Get("cursor")

	var cursor int64
	if cursorStr != "" {
		cursor, _ = strconv.ParseInt(cursorStr, 10, 64)
	}

	slog.Info("relay: instance subscribed", "instance", instanceURL, "cursor", cursor)
	defer slog.Info("relay: instance disconnected", "instance", instanceURL)

	// Register for live events first so nothing is missed during replay.
	// The channel buffer (capacity 256) holds events emitted while we replay.
	ch := s.fanout.Subscribe(instanceURL)
	defer s.fanout.Unsubscribe(instanceURL)

	// Replay missed events from cursor.
	if cursor > 0 {
		if err := s.replaySince(ws, cursor); err != nil {
			slog.Warn("relay: replay error", "err", err)
			return
		}
	}

	// Stream live events until the connection closes. Any events that were
	// emitted during replay are still in the channel buffer and will be
	// delivered here; skip ones already replayed by seq.
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			if evt.Seq <= cursor {
				continue // already replayed
			}
			if err := websocket.JSON.Send(ws, evt); err != nil {
				return
			}
			cursor = evt.Seq
		case <-time.After(30 * time.Second):
			// Send a keepalive ping frame.
			if err := websocket.JSON.Send(ws, map[string]string{"$type": "#ping"}); err != nil {
				return
			}
		}
	}
}

func (s *Server) replaySince(ws *websocket.Conn, cursor int64) error {
	const batchSize = 200
	for {
		events, err := s.store.ListEventsSince(context.Background(), cursor, batchSize)
		if err != nil {
			return err
		}
		for _, evt := range events {
			if err := websocket.JSON.Send(ws, evt); err != nil {
				return err
			}
			cursor = evt.Seq
		}
		if len(events) < batchSize {
			break
		}
	}
	return nil
}
