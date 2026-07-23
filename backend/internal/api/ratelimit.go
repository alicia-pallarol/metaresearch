package api

import (
	"sync"
	"time"
)

// RateLimiter is a per-key token bucket pair: a short burst window and a longer
// hourly window, both of which must have a token available for a write to pass.
//
// State is in-memory, which is the right trade for a single free-tier instance:
// no extra infrastructure, and a restart at worst forgives some quota.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucketPair
	now     func() time.Time

	perMinute  float64
	perHour    float64
	lastSweep  time.Time
	sweepEvery time.Duration
}

type bucketPair struct {
	minuteTokens float64
	hourTokens   float64
	updated      time.Time
}

// NewRateLimiter builds a limiter allowing perMinute writes per minute and
// perHour writes per hour for each key.
func NewRateLimiter(perMinute, perHour int) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*bucketPair),
		now:        time.Now,
		perMinute:  float64(perMinute),
		perHour:    float64(perHour),
		sweepEvery: 10 * time.Minute,
	}
}

// Allow consumes one token for key, reporting whether the request may proceed.
// It is safe for concurrent use.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.sweepLocked(now)

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucketPair{minuteTokens: rl.perMinute, hourTokens: rl.perHour, updated: now}
		rl.buckets[key] = b
	} else {
		elapsed := now.Sub(b.updated).Seconds()
		if elapsed > 0 {
			b.minuteTokens = min(rl.perMinute, b.minuteTokens+elapsed*(rl.perMinute/60))
			b.hourTokens = min(rl.perHour, b.hourTokens+elapsed*(rl.perHour/3600))
			b.updated = now
		}
	}

	if b.minuteTokens < 1 || b.hourTokens < 1 {
		return false
	}
	b.minuteTokens--
	b.hourTokens--
	return true
}

// sweepLocked drops buckets that have refilled completely, so the map cannot grow
// without bound. The caller must hold rl.mu.
func (rl *RateLimiter) sweepLocked(now time.Time) {
	if now.Sub(rl.lastSweep) < rl.sweepEvery {
		return
	}
	rl.lastSweep = now
	for k, b := range rl.buckets {
		if now.Sub(b.updated) > time.Hour {
			delete(rl.buckets, k)
		}
	}
}

// Size reports how many keys are currently tracked (used by tests).
func (rl *RateLimiter) Size() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
}
