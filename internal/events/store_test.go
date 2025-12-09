// ABOUTME: store_test.go contains comprehensive tests for the event store implementation.
// ABOUTME: Tests cover recording, retrieval, persistence, hierarchical filtering, and concurrency.

package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRecordEvent(t *testing.T) {
	// Create temporary event file
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	store, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Create test event
	event := Event{
		ID:        uuid.New().String(),
		AgentID:   "root",
		ParentID:  "",
		Type:      EventAgentStarted,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": "data"},
	}

	// Record event
	if err := store.Record(event); err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	// Verify event was recorded in memory
	events := store.GetEvents("root")
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID != event.ID {
		t.Errorf("Event ID mismatch: expected %s, got %s", event.ID, events[0].ID)
	}

	if events[0].Type != EventAgentStarted {
		t.Errorf("Event type mismatch: expected %s, got %s", EventAgentStarted, events[0].Type)
	}
}

func TestGetEvents_Filtered(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	store, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Create hierarchical events
	events := []Event{
		{
			ID:        uuid.New().String(),
			AgentID:   "root",
			ParentID:  "",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      nil,
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      nil,
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root.1.1",
			ParentID:  "root.1",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      nil,
		},
		{
			ID:        uuid.New().String(),
			AgentID:   "root.2",
			ParentID:  "root",
			Type:      EventAgentStarted,
			Timestamp: time.Now(),
			Data:      nil,
		},
	}

	for _, e := range events {
		if err := store.Record(e); err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Test: Get all events for root (should include all)
	rootEvents := store.GetEvents("root")
	if len(rootEvents) != 4 {
		t.Errorf("Expected 4 events for root tree, got %d", len(rootEvents))
	}

	// Test: Get events for root.1 (should include root.1 and root.1.1)
	root1Events := store.GetEvents("root.1")
	if len(root1Events) != 2 {
		t.Errorf("Expected 2 events for root.1 tree, got %d", len(root1Events))
	}

	// Test: Get events for root.2 (should include only root.2)
	root2Events := store.GetEvents("root.2")
	if len(root2Events) != 1 {
		t.Errorf("Expected 1 event for root.2, got %d", len(root2Events))
	}

	// Test: Get events for root.1.1 (should include only root.1.1)
	root11Events := store.GetEvents("root.1.1")
	if len(root11Events) != 1 {
		t.Errorf("Expected 1 event for root.1.1, got %d", len(root11Events))
	}
}

func TestHierarchicalIDs(t *testing.T) {
	// Reset counters at start of test
	ResetCounters()

	tests := []struct {
		parentID string
		expected string
	}{
		{"", "root"},
		{"root", "root.1"},
		{"root.1", "root.1.1"},
		{"root.1.1", "root.1.1.1"},
	}

	for _, tt := range tests {
		result := GenerateAgentID(tt.parentID)
		if result != tt.expected {
			t.Errorf("GenerateAgentID(%q) = %q, expected %q", tt.parentID, result, tt.expected)
		}
	}

	// Reset for multiple children test
	ResetCounters()

	// Test multiple children
	parent := "root"
	child1 := GenerateAgentID(parent)
	child2 := GenerateAgentID(parent)
	child3 := GenerateAgentID(parent)

	if child1 != "root.1" {
		t.Errorf("First child should be root.1, got %s", child1)
	}
	if child2 != "root.2" {
		t.Errorf("Second child should be root.2, got %s", child2)
	}
	if child3 != "root.3" {
		t.Errorf("Third child should be root.3, got %s", child3)
	}
}

func TestPersistenceToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	// Create store and record events
	store, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	event := Event{
		ID:        uuid.New().String(),
		AgentID:   "root",
		ParentID:  "",
		Type:      EventAgentStarted,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": "persistence"},
	}

	if recordErr := store.Record(event); recordErr != nil {
		t.Fatalf("Failed to record event: %v", recordErr)
	}

	_ = store.Close()

	// Read file and verify JSON Lines format
	content, readErr := os.ReadFile(eventFile)
	if readErr != nil {
		t.Fatalf("Failed to read event file: %v", readErr)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line in file, got %d", len(lines))
	}

	// Parse the JSON
	var parsedEvent Event
	if err := json.Unmarshal([]byte(lines[0]), &parsedEvent); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if parsedEvent.ID != event.ID {
		t.Errorf("Persisted event ID mismatch: expected %s, got %s", event.ID, parsedEvent.ID)
	}

	if parsedEvent.Type != EventAgentStarted {
		t.Errorf("Persisted event type mismatch: expected %s, got %s", EventAgentStarted, parsedEvent.Type)
	}
}

func TestConcurrentRecording(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	store, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Record events concurrently
	numGoroutines := 10
	eventsPerGoroutine := 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := Event{
					ID:        uuid.New().String(),
					AgentID:   "root",
					ParentID:  "",
					Type:      EventStreamChunk,
					Timestamp: time.Now(),
					Data:      map[string]interface{}{"worker": workerID, "iteration": j},
				}
				if err := store.Record(event); err != nil {
					t.Errorf("Worker %d failed to record event: %v", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were recorded
	events := store.GetEvents("root")
	expectedCount := numGoroutines * eventsPerGoroutine
	if len(events) != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, len(events))
	}
}

func TestLoadExistingEvents(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	// Create store and record an event
	store1, err := NewEventStore(eventFile)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	event1 := Event{
		ID:        uuid.New().String(),
		AgentID:   "root",
		ParentID:  "",
		Type:      EventAgentStarted,
		Timestamp: time.Now(),
		Data:      nil,
	}

	if recordErr := store1.Record(event1); recordErr != nil {
		t.Fatalf("Failed to record event: %v", recordErr)
	}
	_ = store1.Close()

	// Create new store pointing to same file
	store2, createErr := NewEventStore(eventFile)
	if createErr != nil {
		t.Fatalf("Failed to create second event store: %v", createErr)
	}
	defer func() { _ = store2.Close() }()

	// Load existing events
	if err := store2.LoadExisting(); err != nil {
		t.Fatalf("Failed to load existing events: %v", err)
	}

	// Verify event was loaded
	events := store2.GetEvents("root")
	if len(events) != 1 {
		t.Fatalf("Expected 1 loaded event, got %d", len(events))
	}

	if events[0].ID != event1.ID {
		t.Errorf("Loaded event ID mismatch: expected %s, got %s", event1.ID, events[0].ID)
	}
}
