// ABOUTME: Shared rate limiter implementation using token bucket algorithm.
// ABOUTME: Prevents multiple parallel agents from causing API rate limit errors.

package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics provides observability into rate limiter behavior.
type Metrics struct {
	TotalAcquired atomic.Int64 // Total successful acquisitions
	TotalWaits    atomic.Int64 // Number of times a goroutine had to wait
	TotalWaitMs   atomic.Int64 // Sum of all wait times in milliseconds
	CurrentTokens atomic.Int64 // Current available tokens
	AvgWaitTimeMs atomic.Int64 // Average wait time in milliseconds (calculated from TotalWaitMs/TotalWaits)
}

// SharedLimiter implements a token bucket rate limiter safe for concurrent use.
// Multiple goroutines can acquire tokens, and the limiter ensures fair distribution
// while respecting the configured rate limits.
type SharedLimiter struct {
	mu         sync.Mutex
	cond       *sync.Cond
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	metrics    Metrics
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewSharedLimiter creates a new rate limiter with the specified parameters.
// maxTokens is the bucket capacity (maximum burst size).
// refillRate is how often a single token is added to the bucket.
// For Anthropic API: maxTokens=50, refillRate=1 minute.
func NewSharedLimiter(maxTokens int, refillRate time.Duration) *SharedLimiter {
	limiter := &SharedLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
		stopChan:   make(chan struct{}),
	}
	limiter.cond = sync.NewCond(&limiter.mu)
	limiter.metrics.CurrentTokens.Store(int64(maxTokens))

	// Start background refill goroutine
	limiter.wg.Add(1)
	go limiter.refillLoop()

	return limiter
}

// Acquire attempts to acquire a single token from the rate limiter.
// Blocks until a token is available or the context is cancelled.
// Returns context.Canceled or context.DeadlineExceeded if context is cancelled/times out.
func (l *SharedLimiter) Acquire(ctx context.Context) error {
	start := time.Now()

	// Set up context cancellation handling
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			// Wake up any waiting goroutines when context is cancelled
			l.cond.Broadcast()
		case <-done:
			// Acquire completed normally
		}
	}()

	l.mu.Lock()
	defer l.mu.Unlock()

	for {
		// Check if limiter is stopped (BEFORE other checks)
		select {
		case <-l.stopChan:
			return fmt.Errorf("rate limiter stopped")
		default:
		}

		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Refill tokens based on elapsed time
		l.refillTokens()

		// Try to acquire token
		if l.tokens > 0 {
			l.tokens--
			l.metrics.TotalAcquired.Add(1)

			waitTime := time.Since(start)
			if waitTime > time.Millisecond {
				l.metrics.TotalWaits.Add(1)
				l.metrics.TotalWaitMs.Add(waitTime.Milliseconds())

				// Calculate true cumulative average
				totalWait := l.metrics.TotalWaitMs.Load()
				waitCount := l.metrics.TotalWaits.Load()
				if waitCount > 0 {
					avgWait := totalWait / waitCount
					l.metrics.AvgWaitTimeMs.Store(avgWait)
				}
			}

			l.metrics.CurrentTokens.Store(int64(l.tokens))
			return nil
		}

		// No tokens available - wait for refill
		// cond.Wait() will unlock the mutex, wait for signal, then relock
		l.cond.Wait()

		// After Wait returns, check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Loop to try acquiring again
		}
	}
}

// refillTokens adds tokens based on elapsed time since last refill.
// Must be called with lock held.
func (l *SharedLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill)
	tokensToAdd := int(elapsed / l.refillRate)

	if tokensToAdd > 0 {
		oldTokens := l.tokens
		l.tokens = min(l.tokens+tokensToAdd, l.maxTokens)
		l.lastRefill = l.lastRefill.Add(time.Duration(tokensToAdd) * l.refillRate)

		// Wake waiting goroutines if we added tokens
		if l.tokens > oldTokens {
			l.cond.Broadcast()
		}
	}
}

// refillLoop runs in a background goroutine to periodically signal waiters.
// This ensures tokens are refilled even if no one is actively trying to acquire.
func (l *SharedLimiter) refillLoop() {
	defer l.wg.Done()

	// Wake up periodically to refill tokens
	ticker := time.NewTicker(l.refillRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			l.refillTokens()
			l.mu.Unlock()
		case <-l.stopChan:
			return
		}
	}
}

// GetMetrics returns a pointer to current metrics.
// Returns pointer to avoid copying sync.Mutex from atomic values.
func (l *SharedLimiter) GetMetrics() *Metrics {
	// Update current tokens in metrics
	l.mu.Lock()
	l.metrics.CurrentTokens.Store(int64(l.tokens))
	l.mu.Unlock()

	// Return pointer (atomic values are safe to read concurrently)
	return &l.metrics
}

// Stop gracefully shuts down the rate limiter.
// Should be called when the limiter is no longer needed.
// Wakes all waiting goroutines which will return an error.
func (l *SharedLimiter) Stop() {
	close(l.stopChan)
	l.cond.Broadcast() // Wake all waiting goroutines
	l.wg.Wait()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
