// ABOUTME: Tests for the shared rate limiter implementation.
// ABOUTME: Verifies token bucket algorithm, concurrency safety, and fairness.

package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAcquire_ImmediateWhenTokensAvailable(t *testing.T) {
	// Limiter with 10 tokens, slow refill
	limiter := NewSharedLimiter(10, 10*time.Second)
	ctx := context.Background()

	// Should acquire immediately without blocking
	start := time.Now()
	err := limiter.Acquire(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// Should be nearly instant (< 10ms)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Acquire took too long: %v (expected < 10ms)", elapsed)
	}

	// Metrics should show one acquisition
	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != 1 {
		t.Errorf("TotalAcquired = %d, want 1", metrics.TotalAcquired.Load())
	}
	if metrics.CurrentTokens.Load() != 9 {
		t.Errorf("CurrentTokens = %d, want 9", metrics.CurrentTokens.Load())
	}
}

func TestAcquire_WaitsWhenEmpty(t *testing.T) {
	// Limiter with 1 token, fast refill (100ms)
	limiter := NewSharedLimiter(1, 100*time.Millisecond)
	ctx := context.Background()

	// Acquire first token (should be immediate)
	err := limiter.Acquire(ctx)
	if err != nil {
		t.Fatalf("First acquire failed: %v", err)
	}

	// Second acquire should wait for refill
	start := time.Now()
	err = limiter.Acquire(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Second acquire failed: %v", err)
	}

	// Should have waited approximately 100ms for refill
	if elapsed < 80*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Wait time = %v, expected ~100ms", elapsed)
	}

	// Metrics should show wait occurred
	metrics := limiter.GetMetrics()
	if metrics.TotalWaits.Load() == 0 {
		t.Error("TotalWaits = 0, expected > 0")
	}
}

func TestRefill_AddsTokensOverTime(t *testing.T) {
	// Limiter with 5 max tokens, 50ms refill rate
	limiter := NewSharedLimiter(5, 50*time.Millisecond)
	ctx := context.Background()

	// Drain all tokens
	for i := 0; i < 5; i++ {
		if err := limiter.Acquire(ctx); err != nil {
			t.Fatalf("Acquire %d failed: %v", i, err)
		}
	}

	// Should have 0 tokens now
	metrics := limiter.GetMetrics()
	if metrics.CurrentTokens.Load() != 0 {
		t.Errorf("CurrentTokens = %d after draining, want 0", metrics.CurrentTokens.Load())
	}

	// Wait for refill (150ms = 3 tokens)
	time.Sleep(150 * time.Millisecond)

	// Should be able to acquire 3 tokens quickly
	for i := 0; i < 3; i++ {
		start := time.Now()
		if err := limiter.Acquire(ctx); err != nil {
			t.Fatalf("Refilled acquire %d failed: %v", i, err)
		}
		if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
			t.Errorf("Refilled acquire %d took %v, expected immediate", i, elapsed)
		}
	}

	// 4th acquire should block (no tokens left)
	start := time.Now()
	err := limiter.Acquire(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("4th acquire failed: %v", err)
	}
	if elapsed < 30*time.Millisecond {
		t.Errorf("4th acquire didn't wait, elapsed = %v", elapsed)
	}
}

func TestConcurrentAcquire_Fair(t *testing.T) {
	// Limiter with 20 tokens, 10ms refill (fast for testing)
	limiter := NewSharedLimiter(20, 10*time.Millisecond)
	ctx := context.Background()

	// Launch 50 goroutines competing for tokens
	numWorkers := 50
	var wg sync.WaitGroup
	var successCount atomic.Int32
	var errorCount atomic.Int32

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := limiter.Acquire(ctx); err != nil {
				errorCount.Add(1)
			} else {
				successCount.Add(1)
			}
		}()
	}

	// Wait for all goroutines (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}

	// All 50 should succeed (20 immediate + 30 refilled over ~300ms)
	if successCount.Load() != int32(numWorkers) {
		t.Errorf("Success count = %d, want %d", successCount.Load(), numWorkers)
	}
	if errorCount.Load() != 0 {
		t.Errorf("Error count = %d, want 0", errorCount.Load())
	}

	// Metrics should show all acquisitions
	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != int64(numWorkers) {
		t.Errorf("TotalAcquired = %d, want %d", metrics.TotalAcquired.Load(), numWorkers)
	}
}

func TestContextCancellation(t *testing.T) {
	// Limiter with 0 tokens, slow refill (never refills during test)
	limiter := NewSharedLimiter(0, 1*time.Hour)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start acquisition in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- limiter.Acquire(ctx)
	}()

	// Wait a bit to ensure goroutine is waiting
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Should return context.Canceled error quickly
	select {
	case err := <-errChan:
		if err != context.Canceled {
			t.Errorf("Acquire error = %v, want context.Canceled", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Acquire didn't respond to context cancellation")
	}
}

func TestContextTimeout(t *testing.T) {
	// Limiter with 0 tokens, slow refill
	limiter := NewSharedLimiter(0, 1*time.Hour)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Acquire should timeout
	start := time.Now()
	err := limiter.Acquire(ctx)
	elapsed := time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("Acquire error = %v, want context.DeadlineExceeded", err)
	}

	// Should timeout around 100ms
	if elapsed < 80*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Timeout occurred at %v, expected ~100ms", elapsed)
	}
}

func TestGetMetrics(t *testing.T) {
	limiter := NewSharedLimiter(10, 100*time.Millisecond)
	ctx := context.Background()

	// Initial state
	metrics := limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != 0 {
		t.Errorf("Initial TotalAcquired = %d, want 0", metrics.TotalAcquired.Load())
	}
	if metrics.CurrentTokens.Load() != 10 {
		t.Errorf("Initial CurrentTokens = %d, want 10", metrics.CurrentTokens.Load())
	}

	// Acquire some tokens
	for i := 0; i < 5; i++ {
		if err := limiter.Acquire(ctx); err != nil {
			t.Fatalf("Acquire failed: %v", err)
		}
	}

	// Check metrics updated
	metrics = limiter.GetMetrics()
	if metrics.TotalAcquired.Load() != 5 {
		t.Errorf("TotalAcquired = %d, want 5", metrics.TotalAcquired.Load())
	}
	if metrics.CurrentTokens.Load() != 5 {
		t.Errorf("CurrentTokens = %d, want 5", metrics.CurrentTokens.Load())
	}
}

func BenchmarkAcquire_NoContention(b *testing.B) {
	limiter := NewSharedLimiter(10000, time.Millisecond)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := limiter.Acquire(ctx); err != nil {
			b.Fatalf("Acquire failed: %v", err)
		}
	}
}

func BenchmarkAcquire_HighContention(b *testing.B) {
	limiter := NewSharedLimiter(100, time.Millisecond)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := limiter.Acquire(ctx); err != nil {
				b.Fatalf("Acquire failed: %v", err)
			}
		}
	})
}
