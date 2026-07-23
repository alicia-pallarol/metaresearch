package api

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-safety-atlas/backend/internal/store"
)

// fakeSummarySource counts calls so tests can assert the cache actually caches.
type fakeSummarySource struct {
	calls atomic.Int64
	mu    sync.Mutex
	value store.Summary
	err   error
	delay time.Duration
}

func (f *fakeSummarySource) Summary(ctx context.Context, iteration int) (store.Summary, error) {
	f.calls.Add(1)
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return store.Summary{}, f.err
	}
	out := f.value
	out.Iteration = iteration
	return out, nil
}

func (f *fakeSummarySource) set(v store.Summary, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.value, f.err = v, err
}

func summaryWithTotal(n int) store.Summary {
	return store.Summary{
		Totals:    store.Totals{TotalSubmissions: n, UniqueSubmitters: n},
		PerArea:   map[string]*store.RatingSummary{},
		PerAgenda: map[string]*store.RatingSummary{},
	}
}

func TestSummaryCacheServesFromMemoryUntilTTL(t *testing.T) {
	src := &fakeSummarySource{}
	src.set(summaryWithTotal(3), nil)

	now := time.Now()
	c := newSummaryCache(src, 0, 45*time.Second)
	c.now = func() time.Time { return now }

	for range 5 {
		got, err := c.Get(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Totals.TotalSubmissions != 3 {
			t.Fatalf("total = %d, want 3", got.Totals.TotalSubmissions)
		}
	}
	if n := src.calls.Load(); n != 1 {
		t.Fatalf("source queried %d times, want 1", n)
	}

	now = now.Add(46 * time.Second)
	if _, err := c.Get(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := src.calls.Load(); n != 2 {
		t.Fatalf("source queried %d times after TTL, want 2", n)
	}
}

func TestSummaryCacheInvalidateForcesRefresh(t *testing.T) {
	src := &fakeSummarySource{}
	src.set(summaryWithTotal(1), nil)

	c := newSummaryCache(src, 0, time.Hour)
	if _, err := c.Get(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	src.set(summaryWithTotal(2), nil)
	c.Invalidate()

	got, err := c.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Totals.TotalSubmissions != 2 {
		t.Fatalf("total = %d after invalidate, want 2", got.Totals.TotalSubmissions)
	}
}

func TestSummaryCacheFallsBackToStaleValueOnError(t *testing.T) {
	src := &fakeSummarySource{}
	src.set(summaryWithTotal(7), nil)

	c := newSummaryCache(src, 0, time.Hour)
	if _, err := c.Get(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	src.set(store.Summary{}, errors.New("database asleep"))
	c.Invalidate()

	got, err := c.Get(context.Background())
	if err != nil {
		t.Fatalf("a stale value should be preferred over an error, got %v", err)
	}
	if got.Totals.TotalSubmissions != 7 {
		t.Fatalf("total = %d, want the stale 7", got.Totals.TotalSubmissions)
	}
}

func TestSummaryCacheReturnsErrorWhenNothingCached(t *testing.T) {
	src := &fakeSummarySource{}
	src.set(store.Summary{}, errors.New("database asleep"))

	c := newSummaryCache(src, 0, time.Hour)
	if _, err := c.Get(context.Background()); err == nil {
		t.Fatal("expected an error when there is no cached value to fall back to")
	}
}

// Run with -race: many readers share one cache, and a write can invalidate it
// while a read is in flight.
func TestSummaryCacheConcurrentReadsAndInvalidation(t *testing.T) {
	src := &fakeSummarySource{delay: time.Millisecond}
	src.set(summaryWithTotal(1), nil)

	c := newSummaryCache(src, 0, 20*time.Millisecond)

	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				if _, err := c.Get(context.Background()); err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}
		}()
	}
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				c.Invalidate()
			}
		}()
	}
	wg.Wait()
}
