// Package atproto provides AT Protocol XRPC client, DID resolution,
// TID generation, and record write helpers for the Sunred API.
package atproto

import (
	"math/rand"
	"sync"
	"time"
)

// base32SortableAlphabet is the AT Proto base32-sortable charset.
// It is a subset of standard base32 with lowercase letters,
// chosen to produce lexicographically sortable strings.
const base32SortableAlphabet = "234567abcdefghijklmnopqrstuvwxyz"

var tidMu sync.Mutex
var lastMicro int64

// NewTID generates an AT Protocol TID (timestamp identifier).
//
// A TID is a 13-character base32-sortable string encoding a 63-bit value:
//   - high 53 bits: microseconds since Unix epoch
//   - low 10 bits: random clock ID (prevents collisions within the same microsecond)
//
// TIDs are monotonically increasing within a process.
func NewTID() string {
	tidMu.Lock()
	now := time.Now().UnixMicro()
	if now <= lastMicro {
		now = lastMicro + 1
	}
	lastMicro = now
	tidMu.Unlock()

	// 10-bit random clock ID
	clockID := rand.Intn(1024) //nolint:gosec // non-crypto, just collision avoidance

	// Pack into a 63-bit value: top 53 bits = timestamp, bottom 10 = clockID
	n := (now << 10) | int64(clockID)

	// Encode as 13 base32-sortable characters (13 * 5 = 65 bits, pad top)
	var buf [13]byte
	for i := 12; i >= 0; i-- {
		buf[i] = base32SortableAlphabet[n&0x1f]
		n >>= 5
	}
	return string(buf[:])
}
