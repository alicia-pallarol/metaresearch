package api

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterBurstAndRefill(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5, 30)
	rl.now = func() time.Time { return now }

	for i := range 5 {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("request %d should have been allowed within the burst", i+1)
		}
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("sixth request in the same minute should have been blocked")
	}

	// A different client is unaffected.
	if !rl.Allow("5.6.7.8") {
		t.Fatal("a different key should have its own bucket")
	}

	// Twelve seconds refills exactly one minute-token (5 per 60s).
	now = now.Add(12 * time.Second)
	if !rl.Allow("1.2.3.4") {
		t.Fatal("a token should have refilled after 12s")
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("only one token should have refilled")
	}
}

func TestRateLimiterHourlyCeiling(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5, 30)
	rl.now = func() time.Time { return now }

	allowed := 0
	// Step a minute at a time so the burst window is never the binding limit.
	for range 20 {
		for range 5 {
			if rl.Allow("1.2.3.4") {
				allowed++
			}
		}
		now = now.Add(time.Minute)
	}

	// 20 minutes buys back a third of the hourly bucket, so the ceiling is the
	// initial 30 plus what refilled, and nowhere near the 100 attempts made.
	if allowed > 41 {
		t.Fatalf("hourly ceiling not enforced: %d requests allowed", allowed)
	}
	if allowed < 30 {
		t.Fatalf("hourly bucket should have allowed at least its initial 30, got %d", allowed)
	}
}

func TestRateLimiterSweepsIdleKeys(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5, 30)
	rl.now = func() time.Time { return now }

	rl.Allow("idle-client")
	if rl.Size() != 1 {
		t.Fatalf("expected 1 tracked key, got %d", rl.Size())
	}

	now = now.Add(2 * time.Hour)
	rl.Allow("fresh-client")

	if rl.Size() != 1 {
		t.Fatalf("idle key should have been swept, tracked keys: %d", rl.Size())
	}
}

// Run with -race: the limiter is shared by every request handler.
func TestRateLimiterConcurrentUse(t *testing.T) {
	rl := NewRateLimiter(5, 30)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := []string{"a", "b", "c"}[i%3]
			for range 10 {
				rl.Allow(key)
				rl.Size()
			}
		}(i)
	}
	wg.Wait()

	if rl.Size() != 3 {
		t.Fatalf("expected 3 tracked keys, got %d", rl.Size())
	}
}
