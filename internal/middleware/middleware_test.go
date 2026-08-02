package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VardaanAggarwal/distlimiter/internal/limiter"
)

func TestMiddleware_RateLimitHeadersAnd429(t *testing.T) {
	l := limiter.NewTokenBucket(2, 1.0) // capacity 2
	mw := New(l)

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Request 1: Allowed (200 OK)
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "user-test")
	rr1 := httptest.NewRecorder()

	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr1.Code)
	}

	if limit := rr1.Header().Get("X-RateLimit-Limit"); limit != "2" {
		t.Errorf("expected X-RateLimit-Limit 2, got %s", limit)
	}

	// Request 2: Allowed (200 OK)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "user-test")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr2.Code)
	}

	// Request 3: Blocked (429 Too Many Requests)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-API-Key", "user-test")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rr3.Code)
	}

	var errResp map[string]interface{}
	json.Unmarshal(rr3.Body.Bytes(), &errResp)

	if errResp["error"] != "Rate limit exceeded" {
		t.Errorf("expected error message 'Rate limit exceeded', got %v", errResp["error"])
	}
}
