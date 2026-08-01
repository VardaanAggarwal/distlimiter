package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	targetURL := flag.String("url", "http://localhost:8080/api/v1/resource", "Target URL")
	totalReqs := flag.Int("n", 10000, "Total requests to send")
	concurrency := flag.Int("c", 50, "Concurrent workers")
	apiKey := flag.String("key", "test-client-1", "API Key header")
	flag.Parse()

	fmt.Printf("=== Running distlimiter Benchmark ===\n")
	fmt.Printf("URL: %s | Reqs: %d | Concurrency: %d\n\n", *targetURL, *totalReqs, *concurrency)

	var (
		allowedCount uint64
		blockedCount uint64
		errorCount   uint64
	)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *concurrency,
		},
		Timeout: 2 * time.Second,
	}

	reqsPerWorker := *totalReqs / *concurrency
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < reqsPerWorker; j++ {
				req, _ := http.NewRequest("GET", *targetURL, nil)
				req.Header.Set("X-API-Key", *apiKey)

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&errorCount, 1)
					continue
				}

				if resp.StatusCode == http.StatusOK {
					atomic.AddUint64(&allowedCount, 1)
				} else if resp.StatusCode == http.StatusTooManyRequests {
					atomic.AddUint64(&blockedCount, 1)
				} else {
					atomic.AddUint64(&errorCount, 1)
				}
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	totalDone := allowedCount + blockedCount + errorCount
	rps := float64(totalDone) / duration.Seconds()

	fmt.Printf("=== Benchmark Results ===\n")
	fmt.Printf("Total Duration: %.3f sec\n", duration.Seconds())
	fmt.Printf("Throughput:     %.2f req/sec\n", rps)
	fmt.Printf("Allowed (200):  %d (%.1f%%)\n", allowedCount, float64(allowedCount)/float64(totalDone)*100)
	fmt.Printf("Blocked (429):  %d (%.1f%%)\n", blockedCount, float64(blockedCount)/float64(totalDone)*100)
	fmt.Printf("Errors:         %d\n", errorCount)
}
