// ABOUTME: Integration tests for multi-agent feature integration
// ABOUTME: Tests event-sourcing, cost tracking, and process registry together

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/events"
	"github.com/2389-research/hex/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiAgentIntegration(t *testing.T) {
	// Set up test environment
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "test_events.jsonl")

	// Initialize event store
	eventStore, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore.Close() }()

	// Initialize cost tracker
	costTracker := cost.NewCostTracker()

	// Set agent ID
	_ = os.Setenv("HEX_AGENT_ID", "test-root")
	defer func() { _ = os.Unsetenv("HEX_AGENT_ID") }()

	// Record test events
	ctx := context.Background()

	startEvent := events.Event{
		ID:        "evt-1",
		AgentID:   "test-root",
		Type:      events.EventType("SessionStart"),
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": true},
	}
	err = eventStore.Record(startEvent)
	require.NoError(t, err)

	// Record cost
	err = costTracker.RecordUsage("test-root", "", "claude-sonnet-4-5-20250929", cost.Usage{
		InputTokens:      1000,
		OutputTokens:     500,
		CacheReadTokens:  0,
		CacheWriteTokens: 0,
	})
	require.NoError(t, err)

	// Close event store
	err = eventStore.Close()
	require.NoError(t, err)

	// Verify event file exists
	_, err = os.Stat(eventFile)
	require.NoError(t, err)

	// Reload and verify events
	eventStore2, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore2.Close() }()

	// Load existing events from file
	err = eventStore2.LoadExisting()
	require.NoError(t, err)

	loadedEvents := eventStore2.GetEvents("test-root")
	assert.Len(t, loadedEvents, 1)
	assert.Equal(t, events.EventType("SessionStart"), loadedEvents[0].Type)

	// Verify cost tracking
	agentCost, err := costTracker.GetAgentCost("test-root")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), agentCost.InputTokens)
	assert.Equal(t, int64(500), agentCost.OutputTokens)
	assert.Greater(t, agentCost.TotalCost, 0.0)

	// Verify process registry (basic check)
	reg := registry.NewProcessRegistry()
	err = reg.Register("test-child", "test-root", &os.Process{Pid: os.Getpid()})
	require.NoError(t, err)

	process := reg.Get("test-child")
	assert.NotNil(t, process)
	assert.Equal(t, "test-root", process.ParentID)

	// Suppress unused variable warning
	_ = ctx
}

func TestEventStoreAndCostTrackerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	eventStore, err := events.NewEventStore(eventFile)
	require.NoError(t, err)

	costTracker := cost.NewCostTracker()

	// Simulate multi-agent scenario
	agents := []string{"root", "root.1", "root.2"}

	for _, agentID := range agents {
		// Record event
		err = eventStore.Record(events.Event{
			ID:        agentID + "-start",
			AgentID:   agentID,
			Type:      events.EventType("AgentStart"),
			Timestamp: time.Now(),
		})
		require.NoError(t, err)

		// Record cost
		err = costTracker.RecordUsage(agentID, "", "claude-sonnet-4-5-20250929", cost.Usage{
			InputTokens:      500,
			OutputTokens:     200,
			CacheReadTokens:  0,
			CacheWriteTokens: 0,
		})
		require.NoError(t, err)
	}

	_ = eventStore.Close()

	// Verify all events recorded
	eventStore2, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore2.Close() }()

	// Load existing events from file
	err = eventStore2.LoadExisting()
	require.NoError(t, err)

	// Verify all events were recorded
	// GetEvents("root") returns root, root.1, root.2 (hierarchical matching)
	rootAndChildren := eventStore2.GetEvents("root")
	assert.Len(t, rootAndChildren, 3, "should have 3 total events (root and 2 children)")

	// Verify specific agents
	root1Events := eventStore2.GetEvents("root.1")
	assert.Len(t, root1Events, 1, "root.1 should have 1 event")

	root2Events := eventStore2.GetEvents("root.2")
	assert.Len(t, root2Events, 1, "root.2 should have 1 event")

	// Verify total cost
	totalCost := 0.0
	for _, agentID := range agents {
		agentCost, err := costTracker.GetAgentCost(agentID)
		require.NoError(t, err)
		totalCost += agentCost.TotalCost
	}
	assert.Greater(t, totalCost, 0.0)
}
