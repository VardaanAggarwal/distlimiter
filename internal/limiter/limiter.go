package limiter

import (
	"time"
)

type Result struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int           `json:"remaining"`
	Limit      int           `json:"limit"`
	ResetAfter time.Duration `json:"reset_after"`
}

type Limiter interface {
	Allow(key string) Result
	AllowN(key string, n int) Result
}
