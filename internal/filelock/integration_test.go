// ABOUTME: Integration tests for file locking with actual file operations
// ABOUTME: Tests concurrent writes, timeouts, and cleanup in realistic scenarios

//go:build !short

package filelock

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestConcurrentFileWrites tests that concurrent writes are properly serialized
func TestConcurrentFileWrites(t *testing.T) {
	manager := NewLockManager()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "concurrent_test.txt")

	numWriters := 10
	var wg sync.WaitGroup
	errors := make(chan error, numWriters)

	// Spawn multiple concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			agentID := "agent-" + string(rune('0'+id))

			// Acquire lock
			if err := manager.Acquire(testFile, agentID, 5*time.Second); err != nil {
				errors <- err
				return
			}
			defer func() { _ = manager.Release(testFile, agentID) }()

			// Read current content
			content, err := os.ReadFile(testFile)
			if err != nil && !os.IsNotExist(err) {
				errors <- err
				return
			}

			// Simulate some work
			time.Sleep(10 * time.Millisecond)

			// Append to content
			newContent := string(content) + "Writer " + string(rune('0'+id)) + "\n"

			// Write back
			if err := os.WriteFile(testFile, []byte(newContent), 0600); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Writer error: %v", err)
	}

	// Verify file has all writes
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read final content: %v", err)
	}

	// Should have exactly numWriters lines
	lines := 0
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}

	if lines != numWriters {
		t.Errorf("Expected %d lines, got %d. Content:\n%s", numWriters, lines, string(content))
	}
}

// TestLockCleanupOnPanic tests that locks are released even if code panics
func TestLockCleanupOnPanic(t *testing.T) {
	manager := NewLockManager()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "panic_test.txt")
	agentID := "panic-agent"

	// Function that panics while holding lock
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic")
			}
		}()

		// Acquire lock
		if err := manager.Acquire(testFile, agentID, 5*time.Second); err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
		defer func() { _ = manager.Release(testFile, agentID) }()

		// Panic!
		panic("test panic")
	}()

	// Lock should be released by defer even after panic
	// Try to acquire again - should succeed immediately
	start := time.Now()
	if err := manager.Acquire(testFile, "other-agent", 1*time.Second); err != nil {
		t.Errorf("Failed to acquire lock after panic: %v", err)
	}
	elapsed := time.Since(start)

	// Should have acquired immediately, not waited for timeout
	if elapsed > 100*time.Millisecond {
		t.Errorf("Lock acquisition took too long: %v (should be immediate)", elapsed)
	}
}

// TestLockManagerStressTest performs stress testing with many concurrent operations
func TestLockManagerStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	manager := NewLockManager()
	tempDir := t.TempDir()
	numFiles := 5
	numAgents := 20
	operationsPerAgent := 50

	var wg sync.WaitGroup
	errors := make(chan error, numAgents*operationsPerAgent)

	// Create test files
	testFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		testFiles[i] = filepath.Join(tempDir, "stress_test_"+string(rune('0'+i))+".txt")
	}

	// Spawn many agents performing many operations
	for a := 0; a < numAgents; a++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()
			agentID := "stress-agent-" + string(rune('0'+agentNum))

			for op := 0; op < operationsPerAgent; op++ {
				// Pick a random file
				fileIdx := op % numFiles
				testFile := testFiles[fileIdx]

				// Acquire lock
				if err := manager.Acquire(testFile, agentID, 2*time.Second); err != nil {
					// Timeout is acceptable in stress test
					continue
				}

				// Read-modify-write
				content, _ := os.ReadFile(testFile)
				newContent := string(content) + "."
				if err := os.WriteFile(testFile, []byte(newContent), 0600); err != nil {
					errors <- err
				}

				// Release lock
				if err := manager.Release(testFile, agentID); err != nil {
					errors <- err
				}

				// Small delay
				time.Sleep(1 * time.Millisecond)
			}
		}(a)
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(30 * time.Second):
		t.Fatal("Stress test timed out - possible deadlock")
	}

	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Operation error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Stress test had %d errors", errorCount)
	}

	// Verify no locks are held at the end
	for _, file := range testFiles {
		if manager.IsLocked(file) {
			t.Errorf("File %s is still locked after stress test", file)
		}
	}
}

// TestLockTimeoutDuringFileWrite tests timeout when file is locked during write
func TestLockTimeoutDuringFileWrite(t *testing.T) {
	manager := NewLockManager()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "timeout_test.txt")

	// Agent 1 acquires lock and holds it
	agent1 := "agent-1"
	if err := manager.Acquire(testFile, agent1, 5*time.Second); err != nil {
		t.Fatalf("Agent 1 failed to acquire lock: %v", err)
	}

	// Agent 2 tries to acquire with short timeout (should fail)
	agent2 := "agent-2"
	start := time.Now()
	err := manager.Acquire(testFile, agent2, 200*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Agent 2 should have timed out")
	}

	if elapsed < 200*time.Millisecond {
		t.Errorf("Timeout occurred too quickly: %v", elapsed)
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("Timeout took too long: %v", elapsed)
	}

	// Release agent 1's lock
	if err := manager.Release(testFile, agent1); err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Now agent 2 should be able to acquire
	if err := manager.Acquire(testFile, agent2, 1*time.Second); err != nil {
		t.Errorf("Agent 2 failed to acquire after agent 1 released: %v", err)
	}
}

// TestReleaseAllOnShutdown simulates cleanup on shutdown
func TestReleaseAllOnShutdown(t *testing.T) {
	manager := NewLockManager()
	tempDir := t.TempDir()
	agentID := "shutdown-agent"

	// Agent acquires multiple locks
	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	for _, file := range files {
		if err := manager.Acquire(file, agentID, 5*time.Second); err != nil {
			t.Fatalf("Failed to acquire lock for %s: %v", file, err)
		}
	}

	// Verify all locked
	for _, file := range files {
		if !manager.IsLocked(file) {
			t.Errorf("File %s should be locked", file)
		}
	}

	// Simulate shutdown - release all locks
	if err := manager.ReleaseAll(agentID); err != nil {
		t.Fatalf("Failed to release all locks: %v", err)
	}

	// Verify all unlocked
	for _, file := range files {
		if manager.IsLocked(file) {
			t.Errorf("File %s should be unlocked after ReleaseAll", file)
		}
	}

	// Other agent should be able to acquire immediately
	otherAgent := "other-agent"
	for _, file := range files {
		start := time.Now()
		if err := manager.Acquire(file, otherAgent, 1*time.Second); err != nil {
			t.Errorf("Failed to acquire %s after ReleaseAll: %v", file, err)
		}
		elapsed := time.Since(start)

		if elapsed > 100*time.Millisecond {
			t.Errorf("Lock acquisition for %s took too long: %v", file, elapsed)
		}
	}
}
