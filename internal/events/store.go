// ABOUTME: store.go implements the EventStore for persisting events to disk in JSON Lines format.
// ABOUTME: Provides thread-safe event recording, retrieval, and hierarchical filtering.

package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// EventStore manages event storage and retrieval
type EventStore struct {
	mu     sync.RWMutex
	events []Event
	file   *os.File
}

var (
	globalStore *EventStore
	globalMu    sync.Mutex
)

// NewEventStore creates a new event store that writes to the specified file
func NewEventStore(filepath string) (*EventStore, error) {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event file: %w", err)
	}

	return &EventStore{
		file:   f,
		events: make([]Event, 0),
	}, nil
}

// LoadExisting loads existing events from the file into memory
func (s *EventStore) LoadExisting() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the file path from the current file
	filePath := s.file.Name()

	// Open file for reading
	readFile, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing file, that's okay
		}
		return fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer func() { _ = readFile.Close() }()

	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("failed to parse event: %w", err)
		}
		s.events = append(s.events, event)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// Record appends an event to the store and persists it to disk
func (s *EventStore) Record(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to in-memory cache
	s.events = append(s.events, event)

	// Write to disk (JSON Lines format)
	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = s.file.Write(append(line, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write event to disk: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for the specified agent and its descendants
func (s *EventStore) GetEvents(agentID string) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Event
	for _, e := range s.events {
		// Include if exact match or child of agentID
		if e.AgentID == agentID || strings.HasPrefix(e.AgentID, agentID+".") {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Close closes the event store file
func (s *EventStore) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// SetGlobal sets the global event store
func SetGlobal(store *EventStore) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalStore = store
}

// Global returns the global event store
func Global() *EventStore {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalStore
}
