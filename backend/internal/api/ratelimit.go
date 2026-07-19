package api

import (
	"sync"
	"time"
)

// rateLimiter is a small in-memory sliding-window limiter keyed by IP hash.
// The Render free tier runs a single instance, so in-memory state is adequate;
// the database count check in the handler is the durable backstop.
type rateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	max    int
	events map[string][]time.Time
	lastGC time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		window: window,
		max:    max,
		events: make(map[string][]time.Time),
		lastGC: time.Now(),
	}
}

// allow records an attempt for key and reports whether it is within the limit.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	if now.Sub(rl.lastGC) > rl.window {
		rl.gc(cutoff)
		rl.lastGC = now
	}

	kept := rl.events[key][:0]
	for _, t := range rl.events[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.max {
		rl.events[key] = kept
		return false
	}
	rl.events[key] = append(kept, now)
	return true
}

func (rl *rateLimiter) gc(cutoff time.Time) {
	for k, ts := range rl.events {
		kept := ts[:0]
		for _, t := range ts {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(rl.events, k)
		} else {
			rl.events[k] = kept
		}
	}
}
