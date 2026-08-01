package middleware

import (
	"encoding/json"
	"fmt"

	"net"
	"net/http"
	"strings"

	"github.com/VardaanAggarwal/distlimiter/internal/limiter"
)

type Middleware struct {
	limiter limiter.Limiter
}

func New(l limiter.Limiter) *Middleware {
	return &Middleware{limiter: l}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := extractClientKey(r)

		result := m.limiter.Allow(key)

		// Set standard RateLimit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%.2f", result.ResetAfter.Seconds()))

		if !result.Allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.ResetAfter.Seconds())+1))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":               "Rate limit exceeded",
				"status":              http.StatusTooManyRequests,
				"retry_after_seconds": int(result.ResetAfter.Seconds()) + 1,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientKey(r *http.Request) string {
	// 1. Check API Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "apikey:" + apiKey
	}

	// 2. Check X-Forwarded-For IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return "ip:" + strings.TrimSpace(ips[0])
	}

	// 3. Fallback to RemoteAddr IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "ip:" + r.RemoteAddr
	}
	return "ip:" + ip
}
