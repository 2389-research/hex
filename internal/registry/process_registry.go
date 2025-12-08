// ABOUTME: Process registry for tracking parent-child agent relationships
// ABOUTME: Implements cascading stop protocol with graceful shutdown

package registry

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// ManagedProcess represents a tracked subprocess with parent-child relationships
type ManagedProcess struct {
	AgentID  string
	ParentID string
	PID      int
	Process  *os.Process
	Children []string // Child agent IDs
}

// ProcessRegistry tracks all managed processes and their relationships
type ProcessRegistry struct {
	mu        sync.RWMutex
	processes map[string]*ManagedProcess
}

var (
	globalRegistry     *ProcessRegistry
	globalRegistryOnce sync.Once
)

// Global returns the singleton process registry instance
func Global() *ProcessRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewProcessRegistry()
	})
	return globalRegistry
}

// NewProcessRegistry creates a new process registry
func NewProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		processes: make(map[string]*ManagedProcess),
	}
}

// Register adds a process to the registry with parent-child relationship
func (r *ProcessRegistry) Register(agentID, parentID string, process *os.Process) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate ID
	if _, exists := r.processes[agentID]; exists {
		return fmt.Errorf("process with ID %s already registered", agentID)
	}

	// Create managed process
	mp := &ManagedProcess{
		AgentID:  agentID,
		ParentID: parentID,
		PID:      process.Pid,
		Process:  process,
		Children: []string{},
	}

	r.processes[agentID] = mp

	// Add to parent's children list
	if parentID != "" {
		if parent, exists := r.processes[parentID]; exists {
			parent.Children = append(parent.Children, agentID)
		}
		// Note: If parent doesn't exist yet, that's ok - might be race condition
		// The orphan detection will handle this case
	}

	return nil
}

// Get retrieves a managed process by agent ID
func (r *ProcessRegistry) Get(agentID string) *ManagedProcess {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processes[agentID]
}

// StopCascading stops a process and all its children recursively
// Children are stopped before parents (depth-first)
// callback is called for each process stopped (optional, for testing)
func (r *ProcessRegistry) StopCascading(agentID string, callback func(string)) error {
	// Snapshot children while holding lock
	r.mu.RLock()
	mp, exists := r.processes[agentID]
	if !exists {
		r.mu.RUnlock()
		return fmt.Errorf("process %s not found in registry", agentID)
	}
	childrenSnapshot := append([]string(nil), mp.Children...) // Copy slice
	r.mu.RUnlock()

	// Recursively stop children (using snapshot, not live data)
	for _, childID := range childrenSnapshot {
		if err := r.StopCascading(childID, callback); err != nil {
			// Log but continue - try to stop as many as possible
			fmt.Fprintf(os.Stderr, "Warning: failed to stop child %s: %v\n", childID, err)
		}
	}

	// Re-acquire lock to check if still exists
	r.mu.RLock()
	mp, stillExists := r.processes[agentID]
	if !stillExists {
		r.mu.RUnlock()
		return nil // Already removed by another goroutine
	}
	r.mu.RUnlock()

	// Stop process (don't hold lock during blocking operation)
	if err := r.stopProcess(mp); err != nil {
		return fmt.Errorf("failed to stop process %s (PID %d): %w", agentID, mp.PID, err)
	}

	// Call callback if provided
	if callback != nil {
		callback(agentID)
	}

	// Remove from registry
	r.mu.Lock()
	delete(r.processes, agentID)
	r.mu.Unlock()

	return nil
}

// stopProcess stops a single process gracefully (SIGTERM then SIGKILL)
func (r *ProcessRegistry) stopProcess(mp *ManagedProcess) error {
	// Try graceful shutdown first (SIGTERM)
	if err := mp.Process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		return nil
	}

	// Wait up to 1 second for graceful shutdown
	done := make(chan error, 1)
	go func() {
		_, err := mp.Process.Wait()
		done <- err
	}()

	select {
	case <-done:
		// Process exited gracefully
		return nil
	case <-time.After(1 * time.Second):
		// Timeout - force kill
		if err := mp.Process.Kill(); err != nil {
			// Process might already be dead
			return nil
		}
		// Wait for kill to complete
		_, _ = mp.Process.Wait()
		return nil
	}
}

// GetOrphans returns agent IDs of processes whose parents are not in the registry
func (r *ProcessRegistry) GetOrphans() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orphans := []string{}
	for agentID, mp := range r.processes {
		if mp.ParentID == "" {
			// Root process, not an orphan
			continue
		}

		// Check if parent exists
		if _, exists := r.processes[mp.ParentID]; !exists {
			orphans = append(orphans, agentID)
		}
	}

	return orphans
}

// Deregister removes a process from the registry
func (r *ProcessRegistry) Deregister(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.processes, agentID)
}

// remove removes a process from registry (internal, for testing)
func (r *ProcessRegistry) remove(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.processes, agentID)
}
