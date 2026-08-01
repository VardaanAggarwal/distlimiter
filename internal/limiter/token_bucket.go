package limiter

import (
	"math"
	"sync"
	"time"
)

type TokenBucketLimiter struct {
	mu         sync.Mutex
	capacity   int     // max burst capacity
	refillRate float64 // tokens per second
	buckets    map[string]*bucket
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*bucket),
	}
}

func (tbl *TokenBucketLimiter) Allow(key string) Result {
	return tbl.AllowN(key, 1)
}

func (tbl *TokenBucketLimiter) AllowN(key string, n int) Result {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	now := time.Now()
	b, exists := tbl.buckets[key]

	if !exists {
		b = &bucket{
			tokens:     float64(tbl.capacity),
			lastRefill: now,
		}
		tbl.buckets[key] = b
	} else {
		// Calculate tokens added since last refill
		elapsed := now.Sub(b.lastRefill).Seconds()
		b.tokens = math.Min(float64(tbl.capacity), b.tokens+(elapsed*tbl.refillRate))
		b.lastRefill = now
	}

	needed := float64(n)
	if b.tokens >= needed {
		b.tokens -= needed
		remaining := int(b.tokens)
		return Result{
			Allowed:    true,
			Remaining:  remaining,
			Limit:      tbl.capacity,
			ResetAfter: 0,
		}
	}

	// Calculate time until next token is available
	missing := needed - b.tokens
	waitSecs := missing / tbl.refillRate
	resetAfter := time.Duration(waitSecs * float64(time.Second))

	return Result{
		Allowed:    false,
		Remaining:  int(b.tokens),
		Limit:      tbl.capacity,
		ResetAfter: resetAfter,
	}
}
