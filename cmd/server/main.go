package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/VardaanAggarwal/distlimiter/internal/limiter"
	"github.com/VardaanAggarwal/distlimiter/internal/middleware"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	algo := flag.String("algo", "token_bucket", "Algorithm: token_bucket or sliding_window")
	rateLimit := flag.Int("limit", 100, "Rate limit capacity")
	flag.Parse()

	var l limiter.Limiter
	if *algo == "sliding_window" {
		l = limiter.NewSlidingWindow(*rateLimit, 1*time.Minute)
		log.Printf("Using Sliding Window algorithm (limit: %d req/min)", *rateLimit)
	} else {
		l = limiter.NewTokenBucket(*rateLimit, 10.0) // 10 tokens/sec refill
		log.Printf("Using Token Bucket algorithm (capacity: %d, refill: 10/s)", *rateLimit)
	}

	mw := middleware.New(l)

	mux := http.NewServeMux()

	// Protected API endpoints
	mux.Handle("/api/v1/resource", mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Protected resource accessed successfully!",
		})
	})))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("distlimiter Gateway running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
