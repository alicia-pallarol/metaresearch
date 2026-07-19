package api

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToMax(t *testing.T) {
	rl := newRateLimiter(3, time.Hour)
	for i := 0; i < 3; i++ {
		if !rl.allow("k") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if rl.allow("k") {
		t.Fatalf("4th attempt should be blocked")
	}
	// Different key is independent.
	if !rl.allow("other") {
		t.Fatalf("independent key should be allowed")
	}
}

func TestRateLimiterExpiresWindow(t *testing.T) {
	rl := newRateLimiter(1, 20*time.Millisecond)
	if !rl.allow("k") {
		t.Fatal("first allowed")
	}
	if rl.allow("k") {
		t.Fatal("second should be blocked")
	}
	time.Sleep(30 * time.Millisecond)
	if !rl.allow("k") {
		t.Fatal("should be allowed after window expires")
	}
}
