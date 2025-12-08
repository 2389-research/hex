// ABOUTME: tracker_test.go implements comprehensive tests for the cost tracking system
// ABOUTME: Tests cover usage recording, tree cost calculation, budget enforcement, and concurrency

package cost

import (
	"os"
	"sync"
	"testing"
)

func TestRecordUsage(t *testing.T) {
	tracker := NewCostTracker()

	usage := Usage{
		InputTokens:      1000,
		OutputTokens:     500,
		CacheReadTokens:  200,
		CacheWriteTokens: 100,
	}

	err := tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	cost, err := tracker.GetAgentCost("agent-1")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	if cost.InputTokens != 1000 {
		t.Errorf("Expected InputTokens=1000, got %d", cost.InputTokens)
	}
	if cost.OutputTokens != 500 {
		t.Errorf("Expected OutputTokens=500, got %d", cost.OutputTokens)
	}
	if cost.CacheReads != 200 {
		t.Errorf("Expected CacheReads=200, got %d", cost.CacheReads)
	}
	if cost.CacheWrites != 100 {
		t.Errorf("Expected CacheWrites=100, got %d", cost.CacheWrites)
	}

	// Record more usage for same agent
	usage2 := Usage{
		InputTokens:  500,
		OutputTokens: 250,
	}
	err = tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage2)
	if err != nil {
		t.Fatalf("RecordUsage second call failed: %v", err)
	}

	cost, err = tracker.GetAgentCost("agent-1")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	// Should be cumulative
	if cost.InputTokens != 1500 {
		t.Errorf("Expected cumulative InputTokens=1500, got %d", cost.InputTokens)
	}
	if cost.OutputTokens != 750 {
		t.Errorf("Expected cumulative OutputTokens=750, got %d", cost.OutputTokens)
	}
}

func TestCalculateCost_Accurate(t *testing.T) {
	tracker := NewCostTracker()

	// Test with exact pricing for claude-sonnet-4-5-20250929
	// Input: $3.00 per 1M tokens
	// Output: $15.00 per 1M tokens
	// Cache Read: $0.30 per 1M tokens
	// Cache Write: $3.75 per 1M tokens

	usage := Usage{
		InputTokens:      1_000_000, // $3.00
		OutputTokens:     1_000_000, // $15.00
		CacheReadTokens:  1_000_000, // $0.30
		CacheWriteTokens: 1_000_000, // $3.75
	}

	err := tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	cost, err := tracker.GetAgentCost("agent-1")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	expectedInputCost := 3.00
	expectedOutputCost := 15.00
	expectedCacheCost := 0.30 + 3.75 // 4.05
	expectedTotal := 22.05

	if cost.InputCost != expectedInputCost {
		t.Errorf("Expected InputCost=$%.2f, got $%.2f", expectedInputCost, cost.InputCost)
	}
	if cost.OutputCost != expectedOutputCost {
		t.Errorf("Expected OutputCost=$%.2f, got $%.2f", expectedOutputCost, cost.OutputCost)
	}
	if cost.CacheCost != expectedCacheCost {
		t.Errorf("Expected CacheCost=$%.2f, got $%.2f", expectedCacheCost, cost.CacheCost)
	}
	if cost.TotalCost != expectedTotal {
		t.Errorf("Expected TotalCost=$%.2f, got $%.2f", expectedTotal, cost.TotalCost)
	}
}

func TestGetTreeCost_Recursive(t *testing.T) {
	tracker := NewCostTracker()

	// Create a tree:
	//   root (agent-1): $10
	//     ├─ child-1 (agent-2): $5
	//     │   └─ grandchild (agent-4): $2
	//     └─ child-2 (agent-3): $3
	// Expected tree cost: $20

	usage1 := Usage{InputTokens: 1_000_000, OutputTokens: 466_667} // ~$10
	err := tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage1)
	if err != nil {
		t.Fatalf("RecordUsage agent-1 failed: %v", err)
	}

	usage2 := Usage{InputTokens: 500_000, OutputTokens: 233_333} // ~$5
	err = tracker.RecordUsage("agent-2", "agent-1", "claude-sonnet-4-5-20250929", usage2)
	if err != nil {
		t.Fatalf("RecordUsage agent-2 failed: %v", err)
	}

	usage3 := Usage{InputTokens: 300_000, OutputTokens: 140_000} // ~$3
	err = tracker.RecordUsage("agent-3", "agent-1", "claude-sonnet-4-5-20250929", usage3)
	if err != nil {
		t.Fatalf("RecordUsage agent-3 failed: %v", err)
	}

	usage4 := Usage{InputTokens: 200_000, OutputTokens: 93_333} // ~$2
	err = tracker.RecordUsage("agent-4", "agent-2", "claude-sonnet-4-5-20250929", usage4)
	if err != nil {
		t.Fatalf("RecordUsage agent-4 failed: %v", err)
	}

	treeCost, err := tracker.GetTreeCost("agent-1")
	if err != nil {
		t.Fatalf("GetTreeCost failed: %v", err)
	}

	// Verify it's close to $20 (allow small floating point error)
	expected := 20.0
	tolerance := 0.5 // $0.50 tolerance
	if treeCost < expected-tolerance || treeCost > expected+tolerance {
		t.Errorf("Expected tree cost ~$%.2f, got $%.2f", expected, treeCost)
	}

	// Verify individual costs
	cost1, _ := tracker.GetAgentCost("agent-1")
	if cost1.TotalCost < 9.5 || cost1.TotalCost > 10.5 {
		t.Errorf("Expected agent-1 cost ~$10, got $%.2f", cost1.TotalCost)
	}

	// Verify subtree cost (agent-2 + agent-4)
	subtreeCost, err := tracker.GetTreeCost("agent-2")
	if err != nil {
		t.Fatalf("GetTreeCost agent-2 failed: %v", err)
	}
	if subtreeCost < 6.5 || subtreeCost > 7.5 {
		t.Errorf("Expected agent-2 subtree cost ~$7 (5+2), got $%.2f", subtreeCost)
	}
}

func TestBudgetEnforcement(t *testing.T) {
	tracker := NewCostTracker()
	enforcer := NewBudgetEnforcer(tracker, 10.0) // $10 max

	// Record usage under budget
	usage1 := Usage{InputTokens: 500_000, OutputTokens: 233_333} // ~$5
	err := tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage1)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	// Check budget - should pass
	err = enforcer.CheckBudget("agent-1")
	if err != nil {
		t.Errorf("CheckBudget should pass for under-budget agent, got error: %v", err)
	}

	// Record more usage to exceed budget
	usage2 := Usage{InputTokens: 1_000_000, OutputTokens: 466_667} // ~$10 more
	err = tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage2)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	// Check budget - should fail
	err = enforcer.CheckBudget("agent-1")
	if err == nil {
		t.Error("CheckBudget should fail for over-budget agent")
	}

	cost, _ := tracker.GetAgentCost("agent-1")
	t.Logf("Agent cost: $%.2f (budget: $10.00)", cost.TotalCost)
}

func TestConcurrentRecording(t *testing.T) {
	tracker := NewCostTracker()

	// Record usage from 10 goroutines concurrently
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()
			agentID := "agent-concurrent"

			for j := 0; j < iterations; j++ {
				usage := Usage{
					InputTokens:  100,
					OutputTokens: 50,
				}
				err := tracker.RecordUsage(agentID, "", "claude-sonnet-4-5-20250929", usage)
				if err != nil {
					t.Errorf("RecordUsage failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify total tokens
	cost, err := tracker.GetAgentCost("agent-concurrent")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	expectedInput := int64(100 * 10 * iterations)
	expectedOutput := int64(50 * 10 * iterations)

	if cost.InputTokens != expectedInput {
		t.Errorf("Expected InputTokens=%d, got %d (data race?)", expectedInput, cost.InputTokens)
	}
	if cost.OutputTokens != expectedOutput {
		t.Errorf("Expected OutputTokens=%d, got %d (data race?)", expectedOutput, cost.OutputTokens)
	}
}

func TestGetAllCosts(t *testing.T) {
	tracker := NewCostTracker()

	// Record for multiple agents
	for i := 1; i <= 3; i++ {
		agentID := "agent-" + string(rune('0'+i))
		usage := Usage{InputTokens: 1000 * i, OutputTokens: 500 * i}
		err := tracker.RecordUsage(agentID, "", "claude-sonnet-4-5-20250929", usage)
		if err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}
	}

	allCosts := tracker.GetAllCosts()
	if len(allCosts) != 3 {
		t.Errorf("Expected 3 cost entries, got %d", len(allCosts))
	}

	// Verify all costs are present
	found := make(map[string]bool)
	for _, cost := range allCosts {
		found[cost.AgentID] = true
	}

	for i := 1; i <= 3; i++ {
		agentID := "agent-" + string(rune('0'+i))
		if !found[agentID] {
			t.Errorf("Agent %s not found in GetAllCosts", agentID)
		}
	}
}

func TestGlobalSingleton(t *testing.T) {
	// Reset global instance
	globalTracker = nil

	tracker1 := Global()
	tracker2 := Global()

	if tracker1 != tracker2 {
		t.Error("Global() should return same instance")
	}

	// Record usage and verify it's accessible from both references
	usage := Usage{InputTokens: 1000, OutputTokens: 500}
	err := tracker1.RecordUsage("test-agent", "", "claude-sonnet-4-5-20250929", usage)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	cost, err := tracker2.GetAgentCost("test-agent")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	if cost.InputTokens != 1000 {
		t.Error("Global singleton not working correctly")
	}
}

func TestUnknownModel(t *testing.T) {
	tracker := NewCostTracker()

	usage := Usage{InputTokens: 1000, OutputTokens: 500}
	err := tracker.RecordUsage("agent-1", "", "unknown-model", usage)

	if err == nil {
		t.Error("Expected error for unknown model, got nil")
	}
}

func TestMarkComplete(t *testing.T) {
	tracker := NewCostTracker()

	usage := Usage{InputTokens: 1000, OutputTokens: 500}
	err := tracker.RecordUsage("agent-1", "", "claude-sonnet-4-5-20250929", usage)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	// Mark as complete
	err = tracker.MarkComplete("agent-1")
	if err != nil {
		t.Fatalf("MarkComplete failed: %v", err)
	}

	cost, err := tracker.GetAgentCost("agent-1")
	if err != nil {
		t.Fatalf("GetAgentCost failed: %v", err)
	}

	if cost.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}

	if cost.CompletedAt.Before(cost.StartedAt) {
		t.Error("CompletedAt should be after StartedAt")
	}
}

func TestMain(m *testing.M) {
	// Reset global tracker before tests
	globalTracker = nil
	os.Exit(m.Run())
}
