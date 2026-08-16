package atproto

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockPDS starts a test HTTP server simulating an AT Proto PDS.
func mockPDS(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestClient_Query(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.test.query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("param") != "value" {
			http.Error(w, "missing param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "test-token")
	var out struct {
		Result string `json:"result"`
	}
	err := client.Query(context.Background(), "com.test.query", map[string]string{"param": "value"}, &out)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if out.Result != "ok" {
		t.Errorf("got result=%q, want 'ok'", out.Result)
	}
}

func TestClient_Query_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.test.bad", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"InvalidInput","message":"missing field"}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "")
	err := client.Query(context.Background(), "com.test.bad", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Procedure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.test.proc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "val" {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"done":true}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "mytoken")
	var out struct {
		Done bool `json:"done"`
	}
	err := client.Procedure(context.Background(), "com.test.proc", map[string]string{"key": "val"}, &out)
	if err != nil {
		t.Fatalf("Procedure: %v", err)
	}
	if !out.Done {
		t.Error("expected done=true")
	}
}

func TestClient_PutRecord(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.atproto.repo.putRecord", func(w http.ResponseWriter, r *http.Request) {
		var body PutRecordInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body.Repo != "did:plc:test" || body.Collection != "io.planetary.graph.follow" || body.Rkey != "rkey123" {
			http.Error(w, "bad input", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uri":"at://did:plc:test/io.planetary.graph.follow/rkey123","cid":"bafy123"}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "tok")
	out, err := client.PutRecord(context.Background(), "did:plc:test", "io.planetary.graph.follow", "rkey123", map[string]string{"subject": "did:plc:other"})
	if err != nil {
		t.Fatalf("PutRecord: %v", err)
	}
	if out.URI == "" {
		t.Error("expected non-empty URI")
	}
}

func TestClient_DeleteRecord(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.atproto.repo.deleteRecord", func(w http.ResponseWriter, r *http.Request) {
		var body DeleteRecordInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "tok")
	err := client.DeleteRecord(context.Background(), "did:plc:test", "io.planetary.graph.follow", "rkey123")
	if err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
}

func TestClient_CreateSession(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.atproto.server.createSession", func(w http.ResponseWriter, r *http.Request) {
		var in CreateSessionInput
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Identifier != "alice" || in.Password != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"AuthenticationRequired","message":"wrong creds"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessJwt":"acc123","refreshJwt":"ref123","did":"did:plc:alice","handle":"alice"}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "")
	sess, err := client.CreateSession(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.DID != "did:plc:alice" || sess.AccessJwt != "acc123" {
		t.Errorf("unexpected session: %+v", sess)
	}

	// Wrong credentials should return an error.
	_, err = client.CreateSession(context.Background(), "alice", "wrong")
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
}

func TestClient_RefreshSession(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.atproto.server.refreshSession", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer refresh-tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessJwt":"newacc","refreshJwt":"newref","did":"did:plc:alice","handle":"alice"}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "")
	sess, err := client.RefreshSession(context.Background(), "refresh-tok")
	if err != nil {
		t.Fatalf("RefreshSession: %v", err)
	}
	if sess.AccessJwt != "newacc" {
		t.Errorf("expected newacc, got %q", sess.AccessJwt)
	}
}

func TestClient_ListRecords(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.atproto.repo.listRecords", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("repo") == "" || q.Get("collection") == "" {
			http.Error(w, "missing params", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"records":[{"uri":"at://did/col/rkey1","cid":"c1","value":{"foo":"bar"}}],"cursor":""}`))
	})
	srv := mockPDS(t, mux)

	client := NewClient(srv.URL, "tok")
	out, err := client.ListRecords(context.Background(), "did:plc:test", "io.planetary.graph.follow", 50, "")
	if err != nil {
		t.Fatalf("ListRecords: %v", err)
	}
	if len(out.Records) != 1 {
		t.Errorf("expected 1 record, got %d", len(out.Records))
	}
}

func TestNewClient_Timeout(t *testing.T) {
	// Server that hangs — client should time out.
	mux := http.NewServeMux()
	mux.HandleFunc("/xrpc/com.test.slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := mockPDS(t, mux)

	// Use a short-timeout client.
	client := &Client{
		pdsURL:     srv.URL,
		token:      "",
		httpClient: &http.Client{Timeout: 50 * time.Millisecond},
	}
	err := client.Query(context.Background(), "com.test.slow", nil, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
