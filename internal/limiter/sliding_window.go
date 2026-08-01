package limiter

import (
	"math"
	"sync"
	"time"
)

type SlidingWindowLimiter struct {
	mu         sync.Mutex
	windowSize time.Duration
	limit      int
	windows    map[string]*userWindow
}

type userWindow struct {
	currWindowID int64
	currCount    int
	prevCount    int
}

func NewSlidingWindow(limit int, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windowSize: windowSize,
		limit:      limit,
		windows:    make(map[string]*userWindow),
	}
}

func (swl *SlidingWindowLimiter) Allow(key string) Result {
	return swl.AllowN(key, 1)
}

func (swl *SlidingWindowLimiter) AllowN(key string, n int) Result {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	currWindowID := now.UnixNano() / swl.windowSize.Nanoseconds()

	uw, exists := swl.windows[key]
	if !exists {
		uw = &userWindow{
			currWindowID: currWindowID,
			currCount:    0,
			prevCount:    0,
		}
		swl.windows[key] = uw
	} else if currWindowID > uw.currWindowID {
		// If more than 1 window has elapsed, previous count is 0
		if currWindowID-uw.currWindowID == 1 {
			uw.prevCount = uw.currCount
		} else {
			uw.prevCount = 0
		}
		uw.currCount = 0
		uw.currWindowID = currWindowID
	}

	// Calculate weighted request count based on elapsed overlap
	windowStart := time.Unix(0, currWindowID*swl.windowSize.Nanoseconds())
	timeIntoCurrWindow := now.Sub(windowStart)
	weight := 1.0 - (float64(timeIntoCurrWindow) / float64(swl.windowSize))

	estimatedCount := float64(uw.currCount) + float64(uw.prevCount)*weight

	if int(estimatedCount)+n <= swl.limit {
		uw.currCount += n
		remaining := swl.limit - (int(estimatedCount) + n)
		return Result{
			Allowed:    true,
			Remaining:  remaining,
			Limit:      swl.limit,
			ResetAfter: swl.windowSize - timeIntoCurrWindow,
		}
	}

	return Result{
		Allowed:    false,
		Remaining:  int(math.Max(0, float64(swl.limit-int(estimatedCount)))),
		Limit:      swl.limit,
		ResetAfter: swl.windowSize - timeIntoCurrWindow,
	}
}
