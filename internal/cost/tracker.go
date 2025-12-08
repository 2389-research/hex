// ABOUTME: tracker.go implements hierarchical cost tracking for multi-agent systems
// ABOUTME: Tracks per-agent costs and calculates recursive tree costs for agent hierarchies

package cost

import (
	"fmt"
	"sync"
	"time"
)

// Usage represents API usage metrics (mirrors core.Usage to avoid import cycle)
type Usage struct {
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
}

// CostTracker tracks API costs for multiple agents
type CostTracker struct {
	mu    sync.RWMutex
	costs map[string]*AgentCost
}

// AgentCost represents the cost and usage for a single agent
type AgentCost struct {
	AgentID      string
	ParentID     string
	Model        string
	InputTokens  int64
	OutputTokens int64
	CacheReads   int64
	CacheWrites  int64
	InputCost    float64
	OutputCost   float64
	CacheCost    float64
	TotalCost    float64
	StartedAt    time.Time
	CompletedAt  *time.Time
}

var (
	globalTracker *CostTracker
	globalMu      sync.Mutex
)

// NewCostTracker creates a new cost tracker
func NewCostTracker() *CostTracker {
	return &CostTracker{
		costs: make(map[string]*AgentCost),
	}
}

// Global returns the global cost tracker singleton
func Global() *CostTracker {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalTracker == nil {
		globalTracker = NewCostTracker()
	}
	return globalTracker
}

// RecordUsage records API usage for an agent
func (t *CostTracker) RecordUsage(agentID, parentID, model string, usage Usage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get or create cost entry
	cost, exists := t.costs[agentID]
	if !exists {
		cost = &AgentCost{
			AgentID:   agentID,
			ParentID:  parentID,
			Model:     model,
			StartedAt: time.Now(),
		}
		t.costs[agentID] = cost
	}

	// Update token counts (cumulative)
	cost.InputTokens += int64(usage.InputTokens)
	cost.OutputTokens += int64(usage.OutputTokens)
	cost.CacheReads += int64(usage.CacheReadTokens)
	cost.CacheWrites += int64(usage.CacheWriteTokens)

	// Get pricing
	pricing, err := getPricing(model)
	if err != nil {
		return err
	}

	// Recalculate costs
	cost.InputCost = calculateTokenCost(cost.InputTokens, pricing.InputTokenPrice)
	cost.OutputCost = calculateTokenCost(cost.OutputTokens, pricing.OutputTokenPrice)
	cost.CacheCost = calculateTokenCost(cost.CacheReads, pricing.CacheReadPrice) +
		calculateTokenCost(cost.CacheWrites, pricing.CacheWritePrice)
	cost.TotalCost = cost.InputCost + cost.OutputCost + cost.CacheCost

	return nil
}

// GetAgentCost returns the cost for a specific agent
func (t *CostTracker) GetAgentCost(agentID string) (*AgentCost, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cost, exists := t.costs[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	// Return a copy to prevent external modification
	costCopy := *cost
	return &costCopy, nil
}

// GetTreeCost calculates the total cost for an agent and all its descendants
func (t *CostTracker) GetTreeCost(rootAgentID string) (float64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.getTreeCostRecursive(rootAgentID)
}

// getTreeCostRecursive is the internal recursive implementation
func (t *CostTracker) getTreeCostRecursive(agentID string) (float64, error) {
	var totalCost float64

	// Get this agent's cost
	if cost, exists := t.costs[agentID]; exists {
		totalCost += cost.TotalCost
	}

	// Recursively add children's costs
	for _, cost := range t.costs {
		if cost.ParentID == agentID {
			childCost, err := t.getTreeCostRecursive(cost.AgentID)
			if err != nil {
				return 0, err
			}
			totalCost += childCost
		}
	}

	return totalCost, nil
}

// GetAllCosts returns all cost entries
func (t *CostTracker) GetAllCosts() []*AgentCost {
	t.mu.RLock()
	defer t.mu.RUnlock()

	costs := make([]*AgentCost, 0, len(t.costs))
	for _, cost := range t.costs {
		costCopy := *cost
		costs = append(costs, &costCopy)
	}

	return costs
}

// MarkComplete marks an agent as complete
func (t *CostTracker) MarkComplete(agentID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cost, exists := t.costs[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	now := time.Now()
	cost.CompletedAt = &now

	return nil
}

// GetChildren returns all direct children of an agent
func (t *CostTracker) GetChildren(agentID string) []*AgentCost {
	t.mu.RLock()
	defer t.mu.RUnlock()

	children := make([]*AgentCost, 0)
	for _, cost := range t.costs {
		if cost.ParentID == agentID {
			costCopy := *cost
			children = append(children, &costCopy)
		}
	}

	return children
}
