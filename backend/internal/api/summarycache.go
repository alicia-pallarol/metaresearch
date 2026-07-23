package api

import (
	"context"
	"sync"
	"time"

	"ai-safety-atlas/backend/internal/store"
)

// summarySource is the slice of the store the summary endpoint needs. Keeping it
// an interface lets the cache be tested without a database.
type summarySource interface {
	Summary(ctx context.Context, iteration int) (store.Summary, error)
}

// summaryCache serves GET /api/summary from memory for ttl, so a page that is
// being read by several people at once produces one query rather than many.
//
// A single mutex guards the whole refresh: concurrent callers queue behind one
// in-flight query instead of stampeding the database. At this scale the extra
// latency for the queued callers is preferable to N duplicate queries.
type summaryCache struct {
	mu        sync.Mutex
	src       summarySource
	iteration int
	ttl       time.Duration
	now       func() time.Time

	value     store.Summary
	fetchedAt time.Time
	hasValue  bool
	// forceRefresh survives a failed refresh, so an invalidated cache keeps
	// trying rather than settling back into serving the old value.
	forceRefresh bool
}

func newSummaryCache(src summarySource, iteration int, ttl time.Duration) *summaryCache {
	return &summaryCache{src: src, iteration: iteration, ttl: ttl, now: time.Now}
}

// Get returns the cached summary, refreshing it when stale.
func (c *summaryCache) Get(ctx context.Context) (store.Summary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hasValue && !c.forceRefresh && c.now().Sub(c.fetchedAt) < c.ttl {
		return c.value, nil
	}

	fresh, err := c.src.Summary(ctx, c.iteration)
	if err != nil {
		if c.hasValue {
			// Serving a slightly stale aggregate beats failing the page.
			return c.value, nil
		}
		return store.Summary{}, err
	}

	c.value = fresh
	c.fetchedAt = c.now()
	c.hasValue = true
	c.forceRefresh = false
	return c.value, nil
}

// Invalidate forces the next Get to hit the database. Called after a successful
// write so a contributor sees their own submission reflected immediately.
func (c *summaryCache) Invalidate() {
	c.mu.Lock()
	c.forceRefresh = true
	c.mu.Unlock()
}
