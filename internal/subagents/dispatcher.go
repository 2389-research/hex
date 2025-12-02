// ABOUTME: Parallel dispatch coordinator for multiple subagents
// ABOUTME: Manages concurrent execution of independent subagent tasks

package subagents

import (
	"context"
	"fmt"
	"sync"
)

// Dispatcher coordinates parallel execution of multiple subagents
type Dispatcher struct {
	// Executor is used to run individual subagents
	Executor *Executor

	// MaxConcurrent limits how many subagents can run in parallel
	MaxConcurrent int
}

// NewDispatcher creates a new parallel dispatcher
func NewDispatcher(executor *Executor) *Dispatcher {
	return &Dispatcher{
		Executor:      executor,
		MaxConcurrent: 10, // Reasonable default
	}
}

// DispatchRequest represents a single subagent to dispatch
type DispatchRequest struct {
	ID      string // Unique identifier for this dispatch
	Request *ExecutionRequest
}

// DispatchResult contains the result of a dispatched subagent
type DispatchResult struct {
	ID     string  // Matches the DispatchRequest.ID
	Result *Result // The execution result
	Error  error   // Error if dispatch failed
}

// DispatchParallel executes multiple subagents concurrently and returns all results
// Results are returned in the order they complete, not request order
func (d *Dispatcher) DispatchParallel(ctx context.Context, requests []*DispatchRequest) []*DispatchResult {
	if len(requests) == 0 {
		return []*DispatchResult{}
	}

	// Create result channel
	resultChan := make(chan *DispatchResult, len(requests))

	// Create semaphore to limit concurrency
	sem := make(chan struct{}, d.MaxConcurrent)

	// Launch all subagents
	var wg sync.WaitGroup
	for _, req := range requests {
		wg.Add(1)

		// Capture loop variable
		request := req

		go func() {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute subagent
			result, err := d.Executor.Execute(ctx, request.Request)

			// Send result
			resultChan <- &DispatchResult{
				ID:     request.ID,
				Result: result,
				Error:  err,
			}
		}()
	}

	// Wait for all to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]*DispatchResult, 0, len(requests))
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// DispatchSequential executes subagents one at a time in order
// Useful when order matters or when resources are constrained
func (d *Dispatcher) DispatchSequential(ctx context.Context, requests []*DispatchRequest) []*DispatchResult {
	results := make([]*DispatchResult, 0, len(requests))

	for _, req := range requests {
		result, err := d.Executor.Execute(ctx, req.Request)

		results = append(results, &DispatchResult{
			ID:     req.ID,
			Result: result,
			Error:  err,
		})

		// Stop on context cancellation
		if ctx.Err() != nil {
			break
		}
	}

	return results
}

// DispatchWithAggregation executes subagents and aggregates their results
// The aggregator function is called with all successful results
func (d *Dispatcher) DispatchWithAggregation(
	ctx context.Context,
	requests []*DispatchRequest,
	aggregator func([]*Result) (interface{}, error),
) (interface{}, error) {
	// Execute all subagents in parallel
	dispatchResults := d.DispatchParallel(ctx, requests)

	// Collect successful results
	successfulResults := make([]*Result, 0, len(dispatchResults))
	var errors []error

	for _, dr := range dispatchResults {
		if dr.Error != nil {
			errors = append(errors, fmt.Errorf("subagent %s failed: %w", dr.ID, dr.Error))
			continue
		}

		if dr.Result != nil && dr.Result.Success {
			successfulResults = append(successfulResults, dr.Result)
		} else if dr.Result != nil {
			errors = append(errors, fmt.Errorf("subagent %s: %s", dr.ID, dr.Result.Error))
		}
	}

	// If all failed, return error
	if len(successfulResults) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("all subagents failed: %v", errors)
		}
		return nil, fmt.Errorf("all subagents failed with no results")
	}

	// Call aggregator with successful results
	return aggregator(successfulResults)
}

// WaitForAny waits for the first subagent to complete successfully
// Returns as soon as one subagent succeeds, cancelling others
func (d *Dispatcher) WaitForAny(ctx context.Context, requests []*DispatchRequest) (*DispatchResult, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}

	// Create context that we can cancel
	raceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create result channel
	resultChan := make(chan *DispatchResult, len(requests))

	// Launch all subagents
	var wg sync.WaitGroup
	for _, req := range requests {
		wg.Add(1)

		// Capture loop variable
		request := req

		go func() {
			defer wg.Done()

			// Execute subagent
			result, err := d.Executor.Execute(raceCtx, request.Request)

			// Send result
			resultChan <- &DispatchResult{
				ID:     request.ID,
				Result: result,
				Error:  err,
			}
		}()
	}

	// Wait for first successful result
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Return first successful result
	var lastError error
	for result := range resultChan {
		if result.Error == nil && result.Result != nil && result.Result.Success {
			cancel() // Cancel remaining subagents
			return result, nil
		}
		if result.Error != nil {
			lastError = result.Error
		}
	}

	// All failed
	if lastError != nil {
		return nil, fmt.Errorf("all subagents failed: %w", lastError)
	}
	return nil, fmt.Errorf("all subagents failed with no results")
}

// DispatchBatch executes subagents in batches of a specified size
// Useful for processing large numbers of subagents without overwhelming resources
func (d *Dispatcher) DispatchBatch(ctx context.Context, requests []*DispatchRequest, batchSize int) []*DispatchResult {
	if batchSize <= 0 {
		batchSize = d.MaxConcurrent
	}

	allResults := make([]*DispatchResult, 0, len(requests))

	// Process in batches
	for i := 0; i < len(requests); i += batchSize {
		// Get batch
		end := i + batchSize
		if end > len(requests) {
			end = len(requests)
		}
		batch := requests[i:end]

		// Execute batch
		batchResults := d.DispatchParallel(ctx, batch)
		allResults = append(allResults, batchResults...)

		// Check for cancellation between batches
		if ctx.Err() != nil {
			break
		}
	}

	return allResults
}

// Statistics provides metrics about dispatch operations
type Statistics struct {
	Total      int
	Successful int
	Failed     int
	Errors     []string
}

// CalculateStatistics computes statistics from dispatch results
func CalculateStatistics(results []*DispatchResult) *Statistics {
	stats := &Statistics{
		Total:  len(results),
		Errors: make([]string, 0),
	}

	for _, r := range results {
		if r.Error != nil {
			stats.Failed++
			stats.Errors = append(stats.Errors, r.Error.Error())
		} else if r.Result != nil && r.Result.Success {
			stats.Successful++
		} else {
			stats.Failed++
			if r.Result != nil && r.Result.Error != "" {
				stats.Errors = append(stats.Errors, r.Result.Error)
			}
		}
	}

	return stats
}
