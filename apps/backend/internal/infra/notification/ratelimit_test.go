package notification

import (
	"testing"
	"time"
)

func TestRateLimiterAllowWithinLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(3, time.Hour)

	for i := 0; i < 3; i++ {
		if !rl.Allow("user1:sms") {
			t.Fatalf("expected Allow to return true on attempt %d", i+1)
		}
	}
}

func TestRateLimiterBlocksOverLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(2, time.Hour)

	rl.Allow("user1:sms") // 1
	rl.Allow("user1:sms") // 2

	if rl.Allow("user1:sms") {
		t.Fatal("expected Allow to return false when over limit")
	}
}

func TestRateLimiterIndependentKeys(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(1, time.Hour)

	if !rl.Allow("user1:sms") {
		t.Fatal("user1 should be allowed")
	}
	if !rl.Allow("user2:sms") {
		t.Fatal("user2 should be allowed (different key)")
	}
	if rl.Allow("user1:sms") {
		t.Fatal("user1 should be blocked on second attempt")
	}
}

func TestRateLimiterWindowResets(t *testing.T) {
	t.Parallel()
	// Use a very short window
	rl := NewRateLimiter(1, 10*time.Millisecond)

	rl.Allow("k")
	if rl.Allow("k") {
		t.Fatal("should be blocked within window")
	}

	time.Sleep(15 * time.Millisecond)

	if !rl.Allow("k") {
		t.Fatal("should be allowed after window expires")
	}
}

func TestRateLimiterReset(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(1, time.Hour)
	rl.Allow("k")

	rl.Reset()

	if !rl.Allow("k") {
		t.Fatal("should be allowed after reset")
	}
}
