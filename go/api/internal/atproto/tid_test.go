package atproto

import (
	"strings"
	"testing"
	"time"
)

func TestNewTID_Format(t *testing.T) {
	tid := NewTID()
	if len(tid) != 13 {
		t.Errorf("expected 13-char TID, got %d chars: %q", len(tid), tid)
	}
	for _, c := range tid {
		if !strings.ContainsRune(base32SortableAlphabet, c) {
			t.Errorf("TID %q contains invalid char %q", tid, c)
		}
	}
}

func TestNewTID_Monotonic(t *testing.T) {
	const n = 1000
	tids := make([]string, n)
	for i := range tids {
		tids[i] = NewTID()
	}
	for i := 1; i < n; i++ {
		if tids[i] <= tids[i-1] {
			t.Errorf("TID[%d]=%q <= TID[%d]=%q: not monotonically increasing", i, tids[i], i-1, tids[i-1])
		}
	}
}

func TestNewTID_Unique(t *testing.T) {
	const n = 10000
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		tid := NewTID()
		if seen[tid] {
			t.Fatalf("duplicate TID %q at iteration %d", tid, i)
		}
		seen[tid] = true
	}
}

func TestNewTID_SortableByTime(t *testing.T) {
	// Two TIDs generated with a known gap must sort in the same order as time.
	t1 := NewTID()
	time.Sleep(2 * time.Millisecond)
	t2 := NewTID()
	if t2 <= t1 {
		t.Errorf("later TID %q should be > earlier TID %q", t2, t1)
	}
}

func TestFormatTime(t *testing.T) {
	ts := time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC)
	got := FormatTime(ts)
	want := "2025-01-15T12:30:00Z"
	if got != want {
		t.Errorf("FormatTime = %q, want %q", got, want)
	}
}
