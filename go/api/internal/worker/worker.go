// Package worker implements a fixed-size worker pool that processes feed
// refresh jobs concurrently.
package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/fuegoio/earthed/go/api/internal/reader/processor"
	"github.com/fuegoio/earthed/go/api/internal/store"
)

// Pool is a worker pool that processes feed refresh jobs.
type Pool struct {
	processor *processor.Processor
	jobs      chan *store.Feed
	wg        sync.WaitGroup
}

// New returns a worker pool with the given concurrency.
func New(proc *processor.Processor, concurrency int) *Pool {
	p := &Pool{
		processor: proc,
		jobs:      make(chan *store.Feed, concurrency*2),
	}
	for i := 0; i < concurrency; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// Submit enqueues a feed for processing. Non-blocking: if the queue is full,
// the feed is skipped (it will be picked up on the next scheduler tick).
func (p *Pool) Submit(ctx context.Context, feed *store.Feed) {
	select {
	case p.jobs <- feed:
	default:
		slog.Warn("worker pool queue full, skipping feed", "feed_id", feed.ID)
	case <-ctx.Done():
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for feed := range p.jobs {
		ctx := context.Background()
		if err := p.processor.ProcessFeed(ctx, feed); err != nil {
			slog.Error("process feed", "feed_id", feed.ID, "url", feed.FeedURL, "err", err)
		}
	}
}

// Stop waits for all workers to finish their current jobs.
func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
