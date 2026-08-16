// Package ratelimit provides a simple in-memory token-bucket rate limiter
// keyed by an arbitrary string (typically a client IP address).
//
// This limiter is single-instance only: state is process-local and not
// shared across API replicas. It is sufficient for a single-node deployment
// and for protecting low-volume endpoints (device-flow issue/confirm) from
// brute force and spam. For multi-instance deployments, move the bucket map
// to a shared store (e.g. Redis) with the same Allow signature.
package ratelimit

import (
	"sync"
	"time"
)

// bucket is a token bucket: tokens refill at a fixed rate up to burst.
type bucket struct {
	tokens   float64
	last     time.Time
}

// Limiter is a concurrency-safe, in-memory collection of token buckets
// keyed by string. Buckets are created lazily on first use.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New returns a Limiter ready to use.
func New() *Limiter {
	return &Limiter{buckets: make(map[string]*bucket)}
}

// Allow reports whether key may proceed under a refill rate of `rate`
// tokens per second with a burst capacity of `burst` tokens. On allow,
// one token is consumed.
func (l *Limiter) Allow(key string, rate float64, burst int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(burst), last: now}
		l.buckets[key] = b
	}

	// Refill: add tokens accrued since the last call, capped at burst.
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(float64(burst), b.tokens+elapsed*rate)
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
