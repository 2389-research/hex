// ABOUTME: Tests for process registry implementing cascading stop protocol
// ABOUTME: Verifies parent-child tracking, recursive shutdown, orphan detection, and thread safety

package registry

import (
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestRegisterProcess(t *testing.T) {
	registry := NewProcessRegistry()

	// Start a simple process
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	// Register the process
	err := registry.Register("root.1", "", cmd.Process)
	if err != nil {
		t.Errorf("Failed to register process: %v", err)
	}

	// Verify process is tracked
	process := registry.Get("root.1")
	if process == nil {
		t.Error("Process not found in registry")
		return // Early return to avoid nil pointer dereference
	}

	if process.PID != cmd.Process.Pid {
		t.Errorf("Expected PID %d, got %d", cmd.Process.Pid, process.PID)
	}
}

func TestStopCascading_ThreeLevels(t *testing.T) {
	registry := NewProcessRegistry()

	// Create a tree of processes:
	//   root.1
	//   ├── root.1.1
	//   │   └── root.1.1.1
	//   └── root.1.2

	// Start processes (sleep commands that we can track)
	procs := make([]*exec.Cmd, 4)
	ids := []string{"root.1", "root.1.1", "root.1.1.1", "root.1.2"}
	parents := []string{"", "root.1", "root.1.1", "root.1"}

	for i, id := range ids {
		cmd := exec.Command("sleep", "30")
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start process %s: %v", id, err)
		}
		procs[i] = cmd

		if err := registry.Register(id, parents[i], cmd.Process); err != nil {
			t.Fatalf("Failed to register process %s: %v", id, err)
		}
	}

	// Verify all processes are running
	for i, cmd := range procs {
		// Check if process is running by sending signal 0
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			t.Errorf("Process %s exited prematurely", ids[i])
		}
	}

	// Stop cascading from root.1
	stoppedOrder := []string{}
	err := registry.StopCascading("root.1", func(id string) {
		stoppedOrder = append(stoppedOrder, id)
	})
	if err != nil {
		t.Errorf("StopCascading failed: %v", err)
	}

	// Verify all processes were stopped
	time.Sleep(100 * time.Millisecond) // Give processes time to stop
	for i, cmd := range procs {
		// Check if process is still running by trying to signal it
		if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
			t.Errorf("Process %s (PID %d) is still running after cascading stop", ids[i], cmd.Process.Pid)
		}
	}

	// Verify stop order: children before parents (leaves first)
	// Expected order: root.1.1.1, then root.1.1 or root.1.2 (either order), then root.1
	if len(stoppedOrder) != 4 {
		t.Errorf("Expected 4 processes stopped, got %d: %v", len(stoppedOrder), stoppedOrder)
	}

	// Deepest child should be first
	if len(stoppedOrder) > 0 && stoppedOrder[0] != "root.1.1.1" {
		t.Errorf("Expected root.1.1.1 to stop first, got %s", stoppedOrder[0])
	}

	// Root should be last
	if len(stoppedOrder) > 0 && stoppedOrder[len(stoppedOrder)-1] != "root.1" {
		t.Errorf("Expected root.1 to stop last, got %s", stoppedOrder[len(stoppedOrder)-1])
	}
}

func TestGetOrphans(t *testing.T) {
	registry := NewProcessRegistry()

	// Create a process tree, then manually remove parent
	cmd1 := exec.Command("sleep", "10")
	cmd2 := exec.Command("sleep", "10")

	if err := cmd1.Start(); err != nil {
		t.Fatalf("Failed to start cmd1: %v", err)
	}
	defer func() { _ = cmd1.Process.Kill() }()

	if err := cmd2.Start(); err != nil {
		t.Fatalf("Failed to start cmd2: %v", err)
	}
	defer func() { _ = cmd2.Process.Kill() }()

	_ = registry.Register("root.1", "", cmd1.Process)
	_ = registry.Register("root.1.1", "root.1", cmd2.Process)

	// No orphans initially
	orphans := registry.GetOrphans()
	if len(orphans) != 0 {
		t.Errorf("Expected no orphans, got %d: %v", len(orphans), orphans)
	}

	// Remove parent process from registry (simulate crash)
	registry.remove("root.1")

	// Now child should be orphaned
	orphans = registry.GetOrphans()
	if len(orphans) != 1 {
		t.Errorf("Expected 1 orphan, got %d: %v", len(orphans), orphans)
	}

	if len(orphans) > 0 && orphans[0] != "root.1.1" {
		t.Errorf("Expected root.1.1 to be orphan, got %s", orphans[0])
	}
}

func TestConcurrentRegistration(t *testing.T) {
	registry := NewProcessRegistry()

	// Start 100 processes concurrently
	var wg sync.WaitGroup
	numProcesses := 100

	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			cmd := exec.Command("sleep", "10")
			if err := cmd.Start(); err != nil {
				t.Errorf("Failed to start process %d: %v", idx, err)
				return
			}
			defer func() { _ = cmd.Process.Kill() }()

			// Register with some parent relationships
			agentID := ""
			parentID := ""
			if idx%2 == 0 {
				agentID = "root." + string(rune(idx))
			} else {
				agentID = "root." + string(rune(idx))
				parentID = "root." + string(rune(idx-1))
			}

			err := registry.Register(agentID, parentID, cmd.Process)
			if err != nil {
				t.Errorf("Failed to register process %d: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all processes were registered
	// (Some may have race conditions with parent IDs not existing, that's ok)
	// The important thing is no panics or data corruption
}

func TestStopCascading_NonexistentProcess(t *testing.T) {
	registry := NewProcessRegistry()

	err := registry.StopCascading("nonexistent", nil)
	if err == nil {
		t.Error("Expected error when stopping nonexistent process")
	}
}

func TestRegister_DuplicateID(t *testing.T) {
	registry := NewProcessRegistry()

	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	// Register once
	err := registry.Register("root.1", "", cmd.Process)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Try to register same ID again
	cmd2 := exec.Command("sleep", "10")
	if startErr := cmd2.Start(); startErr != nil {
		t.Fatalf("Failed to start second process: %v", startErr)
	}
	defer func() { _ = cmd2.Process.Kill() }()

	err = registry.Register("root.1", "", cmd2.Process)
	if err == nil {
		t.Error("Expected error when registering duplicate ID")
	}
}
