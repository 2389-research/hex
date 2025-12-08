// ABOUTME: integration_test.go tests the complete event-sourcing workflow end-to-end.
// ABOUTME: Verifies event recording, persistence, and retrieval.

package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventSourcingIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	// Create event store
	store, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Simulate agent lifecycle events
	events := []Event{
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"prompt": "test prompt"},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventStreamStarted,
			Timestamp: time.Now(),
			Data:      nil,
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventToolCallRequested,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"tool_use_id": "tool-1", "tool_name": "read_file"},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventToolCallApproved,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"tool_use_id": "tool-1", "tool_name": "read_file"},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventToolExecutionStart,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"tool_use_id": "tool-1", "tool_name": "read_file"},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"prompt": "subagent prompt"},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      EventAgentStopped,
			Timestamp: time.Now(),
			Data:      nil,
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventToolExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"tool_use_id": "tool-1",
				"tool_name":   "read_file",
				"success":     true,
			},
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventAgentStopped,
			Timestamp: time.Now(),
			Data:      nil,
		},
	}

	// Record all events
	for _, e := range events {
		if err := store.Record(e); err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Verify all events are in memory
	allEvents := store.GetEvents("root")
	if len(allEvents) != 9 {
		t.Errorf("Expected 9 events in root tree, got %d", len(allEvents))
	}

	// Verify subagent events are included
	subagentEvents := store.GetEvents("root.1")
	if len(subagentEvents) != 2 {
		t.Errorf("Expected 2 events for root.1, got %d", len(subagentEvents))
	}

	// Close and reopen to test persistence
	_ = store.Close()

	// Create new store and load existing events
	store2, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create second event store: %v", err)
	}
	defer func() { _ = store2.Close() }()

	if err := store2.LoadExisting(); err != nil {
		t.Fatalf("Failed to load existing events: %v", err)
	}

	// Verify events were persisted and loaded
	loadedEvents := store2.GetEvents("root")
	if len(loadedEvents) != 9 {
		t.Errorf("Expected 9 events after reload, got %d", len(loadedEvents))
	}

	// Verify file exists and has correct number of lines
	content, err := os.ReadFile(eventFile)
	if err != nil {
		t.Fatalf("Failed to read event file: %v", err)
	}

	lines := 0
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}

	if lines != 9 {
		t.Errorf("Expected 9 lines in event file, got %d", lines)
	}
}

func TestAgentIDGeneration(t *testing.T) {
	// Reset counters for clean test
	ResetCounters()

	// Test root agent
	rootID := GenerateAgentID("")
	if rootID != "root" {
		t.Errorf("Root agent ID should be 'root', got %s", rootID)
	}

	// Test child agents
	child1 := GenerateAgentID("root")
	if child1 != "root.1" {
		t.Errorf("First child should be root.1, got %s", child1)
	}

	child2 := GenerateAgentID("root")
	if child2 != "root.2" {
		t.Errorf("Second child should be root.2, got %s", child2)
	}

	// Test nested agents
	grandchild1 := GenerateAgentID("root.1")
	if grandchild1 != "root.1.1" {
		t.Errorf("First grandchild should be root.1.1, got %s", grandchild1)
	}

	grandchild2 := GenerateAgentID("root.1")
	if grandchild2 != "root.1.2" {
		t.Errorf("Second grandchild should be root.1.2, got %s", grandchild2)
	}

	// Test deep nesting
	deepChild := GenerateAgentID("root.1.1")
	if deepChild != "root.1.1.1" {
		t.Errorf("Deep child should be root.1.1.1, got %s", deepChild)
	}
}
