// ABOUTME: Shared registry for managing background bash processes
// ABOUTME: Allows tracking, retrieving, and cleaning up background processes across tools

package tools

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// BackgroundProcess represents a running or completed background bash process
type BackgroundProcess struct {
	ID         string       // Unique identifier for the process
	Command    string       // The command being executed
	StartTime  time.Time    // When the process started
	Stdout     []string     // Lines of stdout output
	Stderr     []string     // Lines of stderr output
	ReadOffset int          // Number of lines already read by BashOutput
	Done       bool         // Whether the process has finished
	ExitCode   int          // Exit code (valid only if Done is true)
	Process    *os.Process  // The actual OS process
	mu         sync.RWMutex // Protects all fields
}

// AppendStdout adds a line to stdout in a thread-safe manner
func (bp *BackgroundProcess) AppendStdout(line string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.Stdout = append(bp.Stdout, line)
}

// AppendStderr adds a line to stderr in a thread-safe manner
func (bp *BackgroundProcess) AppendStderr(line string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.Stderr = append(bp.Stderr, line)
}

// GetNewOutput returns output since the last read and updates the read offset
func (bp *BackgroundProcess) GetNewOutput() (stdout []string, stderr []string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	stdout = make([]string, 0)
	stderr = make([]string, 0)

	// Get stdout lines since last read
	if bp.ReadOffset < len(bp.Stdout) {
		stdout = bp.Stdout[bp.ReadOffset:]
	}

	// For stderr, always include all lines (stderr doesn't use offset tracking)
	stderr = bp.Stderr

	// Update read offset to current stdout length
	bp.ReadOffset = len(bp.Stdout)

	return stdout, stderr
}

// MarkDone marks the process as completed with the given exit code
func (bp *BackgroundProcess) MarkDone(exitCode int) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.Done = true
	bp.ExitCode = exitCode
}

// IsDone returns whether the process has finished
func (bp *BackgroundProcess) IsDone() bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.Done
}

// GetExitCode returns the exit code (only valid if Done is true)
func (bp *BackgroundProcess) GetExitCode() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.ExitCode
}

// BackgroundProcessRegistry stores running background processes
type BackgroundProcessRegistry struct {
	mu        sync.RWMutex
	processes map[string]*BackgroundProcess
}

var (
	// globalBackgroundRegistry is the shared registry for background processes
	globalBackgroundRegistry = &BackgroundProcessRegistry{
		processes: make(map[string]*BackgroundProcess),
	}
)

// GetBackgroundRegistry returns the global background process registry
func GetBackgroundRegistry() *BackgroundProcessRegistry {
	return globalBackgroundRegistry
}

// Register adds a background process to the registry
func (r *BackgroundProcessRegistry) Register(proc *BackgroundProcess) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.processes[proc.ID]; exists {
		return fmt.Errorf("background process %s already exists", proc.ID)
	}

	r.processes[proc.ID] = proc
	return nil
}

// Get retrieves a background process by ID
func (r *BackgroundProcessRegistry) Get(id string) (*BackgroundProcess, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proc, exists := r.processes[id]
	if !exists {
		return nil, fmt.Errorf("background process %s not found", id)
	}

	return proc, nil
}

// Remove removes a background process from the registry
func (r *BackgroundProcessRegistry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.processes[id]; !exists {
		return fmt.Errorf("background process %s not found", id)
	}

	delete(r.processes, id)
	return nil
}

// List returns all background process IDs
func (r *BackgroundProcessRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.processes))
	for id := range r.processes {
		ids = append(ids, id)
	}
	return ids
}

// Legacy functions for backwards compatibility
// These maintain the old API while using the new BackgroundProcess structure

// RegisterBackgroundProcess adds a process to the global registry (legacy)
func RegisterBackgroundProcess(shellID string, process *os.Process) {
	globalBackgroundRegistry.mu.Lock()
	defer globalBackgroundRegistry.mu.Unlock()

	// Create a BackgroundProcess if it doesn't exist
	if _, exists := globalBackgroundRegistry.processes[shellID]; !exists {
		globalBackgroundRegistry.processes[shellID] = &BackgroundProcess{
			ID:        shellID,
			Process:   process,
			StartTime: time.Now(),
			Stdout:    []string{},
			Stderr:    []string{},
		}
	} else {
		// Update existing process
		globalBackgroundRegistry.processes[shellID].Process = process
	}
}

// GetBackgroundProcess retrieves a process from the global registry (legacy)
func GetBackgroundProcess(shellID string) *os.Process {
	globalBackgroundRegistry.mu.RLock()
	defer globalBackgroundRegistry.mu.RUnlock()

	if proc, exists := globalBackgroundRegistry.processes[shellID]; exists {
		return proc.Process
	}
	return nil
}

// UnregisterBackgroundProcess removes a process from the global registry (legacy)
func UnregisterBackgroundProcess(shellID string) {
	globalBackgroundRegistry.mu.Lock()
	defer globalBackgroundRegistry.mu.Unlock()
	delete(globalBackgroundRegistry.processes, shellID)
}

// ListBackgroundProcesses returns a list of all registered shell IDs (legacy)
func ListBackgroundProcesses() []string {
	globalBackgroundRegistry.mu.RLock()
	defer globalBackgroundRegistry.mu.RUnlock()

	ids := make([]string, 0, len(globalBackgroundRegistry.processes))
	for id := range globalBackgroundRegistry.processes {
		ids = append(ids, id)
	}
	return ids
}
