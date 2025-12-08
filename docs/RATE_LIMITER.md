# Shared API Rate Limiter

## Overview

The shared rate limiter prevents multiple parallel agents from overwhelming the Anthropic API and causing 429 (rate limit exceeded) errors. It implements a token bucket algorithm that ensures fair distribution of API calls across all concurrent agents.

## Implementation

### Location
- **Package**: `internal/ratelimit`
- **Main File**: `internal/ratelimit/limiter.go`
- **Tests**: `internal/ratelimit/limiter_test.go`
- **Integration Tests**: `internal/ratelimit/integration_test.go`

### Token Bucket Algorithm

The rate limiter uses a classic token bucket algorithm:

1. **Bucket Capacity**: 50 tokens (matching Anthropic API limits)
2. **Refill Rate**: 1 token per minute
3. **Acquisition**: Each API call consumes 1 token
4. **Blocking**: When no tokens are available, requests wait until refill

```go
// Global rate limiter shared across all clients
var globalLimiter = ratelimit.NewSharedLimiter(50, time.Minute)
```

### Key Features

#### 1. Fair Distribution
Uses `sync.Cond` for fair FIFO distribution of tokens across waiting goroutines.

#### 2. Context Support
Respects context cancellation and timeouts:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := globalLimiter.Acquire(ctx); err != nil {
    // Handle timeout or cancellation
}
```

#### 3. Burst Handling
Full bucket (50 tokens) allows bursts of requests without waiting:
- First 50 requests: immediate
- Subsequent requests: wait for refill

#### 4. Metrics
Provides observability into rate limiting behavior:
```go
type Metrics struct {
    TotalAcquired atomic.Int64 // Total successful acquisitions
    TotalWaits    atomic.Int64 // Number of times goroutines waited
    CurrentTokens atomic.Int64 // Current available tokens
    AvgWaitTimeMs atomic.Int64 // Average wait time in milliseconds
}
```

## Usage

### In API Client

The rate limiter is automatically integrated into the API client:

**CreateMessage** (internal/core/client.go):
```go
func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
    // Acquire rate limit token before making API call
    if err := globalLimiter.Acquire(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }
    // ... make API call
}
```

**CreateMessageStream** (internal/core/stream.go):
```go
func (c *Client) CreateMessageStream(ctx context.Context, req MessageRequest) (<-chan *StreamChunk, error) {
    // Acquire rate limit token before making API call
    if err := globalLimiter.Acquire(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }
    // ... make streaming API call
}
```

### Manual Usage

For custom scenarios:
```go
import "github.com/2389-research/hex/internal/ratelimit"

// Create limiter (usually global)
limiter := ratelimit.NewSharedLimiter(50, time.Minute)
defer limiter.Stop()

// Acquire token before API call
ctx := context.Background()
if err := limiter.Acquire(ctx); err != nil {
    log.Fatalf("Rate limit error: %v", err)
}

// Make API call
// ...

// Check metrics
metrics := limiter.GetMetrics()
fmt.Printf("Total requests: %d, Waits: %d\n",
    metrics.TotalAcquired.Load(),
    metrics.TotalWaits.Load())
```

## Behavior Examples

### Scenario 1: Burst of Requests
```
Time: 0ms
Tokens: 50
Action: 50 parallel agents make requests
Result: All succeed immediately (burst capacity)
Tokens: 0

Time: 60000ms (1 minute)
Tokens: 1 (refilled)
Action: 1 request
Result: Succeeds immediately
Tokens: 0
```

### Scenario 2: Sustained Load
```
Time: 0ms
Tokens: 50
Action: 100 parallel agents make requests
Result: First 50 succeed immediately, next 50 wait

Time: 60000ms
Tokens: 1 (refilled)
Result: 1 waiting agent succeeds
Remaining waiting: 49

Time: 120000ms
Tokens: 1 (refilled)
Result: 1 waiting agent succeeds
Remaining waiting: 48

... continues until all 100 complete over ~50 minutes
```

### Scenario 3: Context Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// If no tokens available within 5 seconds
err := limiter.Acquire(ctx)
// err == context.DeadlineExceeded
```

## Performance

### Benchmarks

```
BenchmarkAcquire_NoContention-12    	 1000000	      1059 ns/op
BenchmarkAcquire_HighContention-12  	  200000	      8432 ns/op
```

### Memory
- Single global instance: ~200 bytes
- No per-request allocations (uses sync.Cond)
- Background goroutine: 1 per limiter instance

### Thread Safety
- All operations are thread-safe
- Uses `sync.Mutex` for state protection
- Atomic operations for metrics

## Testing

### Unit Tests
```bash
go test ./internal/ratelimit/... -v
```

Tests cover:
- Immediate acquisition when tokens available
- Waiting when tokens exhausted
- Token refill over time
- Concurrent acquisition fairness
- Context cancellation
- Context timeout
- Metrics accuracy

### Integration Tests
```bash
go test ./internal/ratelimit/... -v -run TestIntegration
```

Tests cover:
- 20 concurrent agents scenario
- Rate limit preventing 429 errors
- Burst handling (50 immediate requests)

### Example Output
```
=== RUN   TestIntegration_ConcurrentRequests
    integration_test.go:77: 20 agents completed in 1.502s with rate limiting
    integration_test.go:78: Metrics: Acquired=20, Waits=15, AvgWait=1401ms
--- PASS: TestIntegration_ConcurrentRequests (1.50s)
```

## Configuration

### Anthropic API Limits
Per [Anthropic documentation](https://docs.anthropic.com/en/api/rate-limits):
- Tier 1: 50 requests per minute
- Tier 2: 1000 requests per minute
- Tier 3: 2000 requests per minute

**Current Implementation**: Configured for Tier 1 (50/minute)

### Adjusting for Higher Tiers
If you have a higher tier API key, adjust the global limiter:

```go
// For Tier 2 (1000/minute)
var globalLimiter = ratelimit.NewSharedLimiter(1000, time.Minute)

// For Tier 3 (2000/minute)
var globalLimiter = ratelimit.NewSharedLimiter(2000, time.Minute)
```

## Troubleshooting

### Still Getting 429 Errors
1. **Check rate limit tier**: Ensure limiter matches your API tier
2. **Multiple processes**: Each process has its own limiter (not shared across processes)
3. **Other API calls**: Check for API calls outside the limiter

### Slow Performance
1. **Expected behavior**: With 50 requests/minute, expect ~1.2s per request under sustained load
2. **Burst capacity**: First 50 requests are immediate
3. **Context timeouts**: Set appropriate timeouts for your use case

### Deadlocks
1. **Context required**: Always pass a context with timeout
2. **Stop limiter**: Call `limiter.Stop()` when shutting down
3. **Check metrics**: Use `GetMetrics()` to debug wait times

## Future Enhancements

Potential improvements for future iterations:

1. **Per-Tier Configuration**: Auto-detect API tier from responses
2. **Dynamic Adjustment**: Adjust rate based on 429 responses
3. **Distributed Limiting**: Share rate limit across multiple processes
4. **Token Priority**: Priority queues for critical requests
5. **Backoff Strategy**: Exponential backoff on rate limit errors

## Related Files

- `internal/core/client.go` - Integration in CreateMessage
- `internal/core/stream.go` - Integration in CreateMessageStream
- `MULTIAGENT_IMPROVEMENTS_PLAN.md` - Overall multiagent architecture plan
