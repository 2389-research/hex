// ABOUTME: budget.go implements budget enforcement for cost tracking
// ABOUTME: Prevents agents from exceeding maximum cost limits

package cost

import (
	"fmt"
)

// BudgetEnforcer enforces cost limits on agents
type BudgetEnforcer struct {
	tracker *CostTracker
	maxCost float64
}

// NewBudgetEnforcer creates a new budget enforcer
func NewBudgetEnforcer(tracker *CostTracker, maxCost float64) *BudgetEnforcer {
	return &BudgetEnforcer{
		tracker: tracker,
		maxCost: maxCost,
	}
}

// CheckBudget verifies that an agent hasn't exceeded the budget
func (e *BudgetEnforcer) CheckBudget(agentID string) error {
	cost, err := e.tracker.GetAgentCost(agentID)
	if err != nil {
		return err
	}

	if cost.TotalCost > e.maxCost {
		return fmt.Errorf("budget exceeded: $%.2f > $%.2f", cost.TotalCost, e.maxCost)
	}

	return nil
}

// CheckTreeBudget verifies that an agent tree hasn't exceeded the budget
func (e *BudgetEnforcer) CheckTreeBudget(rootAgentID string) error {
	treeCost, err := e.tracker.GetTreeCost(rootAgentID)
	if err != nil {
		return err
	}

	if treeCost > e.maxCost {
		return fmt.Errorf("tree budget exceeded: $%.2f > $%.2f", treeCost, e.maxCost)
	}

	return nil
}

// GetRemainingBudget returns how much budget is remaining
func (e *BudgetEnforcer) GetRemainingBudget(agentID string) (float64, error) {
	cost, err := e.tracker.GetAgentCost(agentID)
	if err != nil {
		return 0, err
	}

	remaining := e.maxCost - cost.TotalCost
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
