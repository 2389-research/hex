// ABOUTME: Integration tests for rate limiter with API client.
// ABOUTME: Verifies rate limiting prevents 429 errors in concurrent scenarios.

package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestIntegration_ConcurrentRequests simulates multiple parallel agents making API calls
func TestIntegration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a limiter with realistic parameters (reduced for faster testing)
	limiter := NewSharedLimiter(5, 100*time.Millisecond)
	defer limiter.Stop()

	// Simulate 20 concurrent agents
	numAgents := 20
	var wg sync.WaitGroup
	ctx := context.Background()

	start := time.Now()
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numAgents; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each agent makes a request
			if err := limiter.Acquire(ctx); err != nil {
				t.Errorf("Agent %d failed to acquire: %v", id, err)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify all agents succeeded
	if successCount != numAgents {
		t.Errorf("Success count = %d, want %d", successCount, numAgents)
	}

	// With 5 tokens initially and 1 token per 100ms:
	// - First 5 agents: immediate (0ms)
	// - Next 15 agents: need refills over ~1.5 seconds
	// Total should be around 1.5 seconds
	expectedMin := 1400 * time.Millisecond
	expectedMax := 2000 * time.Millisecond

	if elapsed < expectedMin {
		t.Errorf("Completed too quickly: %v (expected >= %v)", elapsed, expectedMin)
	}
	if elapsed > expectedMax {
		t.Errorf("Took too long: %v (expected <= %v)", elapsed, expectedMax)
	}

	// Verify metrics
	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != int64(numAgents) {
		t.Errorf("TotalAcquired = %d, want %d", metrics.TotalAcquired.Load(), numAgents)
	}

	t.Logf("20 agents completed in %v with rate limiting", elapsed)
	t.Logf("Metrics: Acquired=%d, Waits=%d, AvgWait=%dms",
		metrics.TotalAcquired.Load(),
		metrics.TotalWaits.Load(),
		metrics.AvgWaitTimeMs.Load())
}

// TestIntegration_RateLimitPrevents429 verifies rate limiting prevents overwhelming API
func TestIntegration_RateLimitPrevents429(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Simulate aggressive rate limit (2 requests per 100ms = 20/sec)
	limiter := NewSharedLimiter(2, 100*time.Millisecond)
	defer limiter.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to make 10 requests
	numRequests := 10
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := limiter.Acquire(ctx); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// All should succeed (just slowly)
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Errorf("Unexpected error: %v", err)
	}

	if errorCount > 0 {
		t.Errorf("Got %d errors, expected 0", errorCount)
	}

	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != int64(numRequests) {
		t.Errorf("TotalAcquired = %d, want %d", metrics.TotalAcquired.Load(), numRequests)
	}
}

// TestIntegration_BurstHandling verifies burst capacity works correctly
func TestIntegration_BurstHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Large bucket (50 tokens) for burst handling
	limiter := NewSharedLimiter(50, time.Minute)
	defer limiter.Stop()

	ctx := context.Background()

	// Burst of 50 requests should all succeed immediately
	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := limiter.Acquire(ctx); err != nil {
				t.Errorf("Burst acquire failed: %v", err)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// All 50 should complete quickly (< 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Burst took %v, expected < 100ms", elapsed)
	}

	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != 50 {
		t.Errorf("TotalAcquired = %d, want 50", metrics.TotalAcquired.Load())
	}
	if metrics.TotalWaits.Load() != 0 {
		t.Errorf("TotalWaits = %d, want 0 (burst should not wait)", metrics.TotalWaits.Load())
	}

	t.Logf("Burst of 50 requests completed in %v", elapsed)
}
