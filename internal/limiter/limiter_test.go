package limiter

import (
	"testing"
	"time"
)

func TestTokenBucket_CapacityBurst(t *testing.T) {
	tb := NewTokenBucket(5, 1.0) // 5 max tokens, 1 token/sec refill

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		res := tb.Allow("user1")
		if !res.Allowed {
			t.Errorf("request %d should be allowed, got false", i+1)
		}
	}

	// 6th request should be rejected
	res := tb.Allow("user1")
	if res.Allowed {
		t.Error("6th request should be rate limited")
	}

	// Wait 1.1s for 1 token to refill
	time.Sleep(1100 * time.Millisecond)

	res = tb.Allow("user1")
	if !res.Allowed {
		t.Error("request after token refill should be allowed")
	}
}

func TestSlidingWindow_Smoothing(t *testing.T) {
	sw := NewSlidingWindow(10, 1*time.Second) // 10 req / sec

	// Send 10 requests
	for i := 0; i < 10; i++ {
		res := sw.Allow("user2")
		if !res.Allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 11th request rejected
	res := sw.Allow("user2")
	if res.Allowed {
		t.Error("11th request should be rejected")
	}
}

func TestLimiter_MultiUserIsolation(t *testing.T) {
	tb := NewTokenBucket(2, 0.5)

	// Exhaust user A
	tb.Allow("userA")
	tb.Allow("userA")

	resA := tb.Allow("userA")
	if resA.Allowed {
		t.Error("userA should be blocked")
	}

	// User B should still have full quota
	resB := tb.Allow("userB")
	if !resB.Allowed {
		t.Error("userB should not be affected by userA quota exhaustion")
	}
}
