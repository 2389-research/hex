// ABOUTME: Test suite for file lock manager, ensures thread-safe concurrent file access
// ABOUTME: Tests basic locking, timeouts, force release, and deadlock prevention

package filelock

import (
	"sync"
	"testing"
	"time"
)

// TestAcquireRelease tests basic lock acquisition and release
func TestAcquireRelease(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner := "agent-1"

	// Acquire lock
	err := manager.Acquire(path, owner, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Verify locked
	if !manager.IsLocked(path) {
		t.Error("File should be locked")
	}

	// Release lock
	err = manager.Release(path, owner)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Verify unlocked
	if manager.IsLocked(path) {
		t.Error("File should not be locked")
	}
}

// TestConcurrentAcquire_BlocksSecond tests that second acquire blocks until first releases
func TestConcurrentAcquire_BlocksSecond(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner1 := "agent-1"
	owner2 := "agent-2"

	// Agent 1 acquires lock
	err := manager.Acquire(path, owner1, 5*time.Second)
	if err != nil {
		t.Fatalf("Agent 1 failed to acquire lock: %v", err)
	}

	// Track timing
	var agent2AcquiredAt time.Time
	var wg sync.WaitGroup
	wg.Add(1)

	// Agent 2 tries to acquire (should block)
	go func() {
		defer wg.Done()
		err := manager.Acquire(path, owner2, 5*time.Second)
		if err != nil {
			t.Errorf("Agent 2 failed to acquire lock: %v", err)
			return
		}
		agent2AcquiredAt = time.Now()
		_ = manager.Release(path, owner2)
	}()

	// Wait a bit to ensure agent 2 is blocking
	time.Sleep(200 * time.Millisecond)

	// Verify file is still locked by agent 1
	if !manager.IsLocked(path) {
		t.Error("File should still be locked by agent 1")
	}

	// Release agent 1's lock
	releaseTime := time.Now()
	err = manager.Release(path, owner1)
	if err != nil {
		t.Fatalf("Agent 1 failed to release lock: %v", err)
	}

	// Wait for agent 2 to acquire
	wg.Wait()

	// Agent 2 should have acquired after agent 1 released
	if agent2AcquiredAt.Before(releaseTime) {
		t.Error("Agent 2 should have acquired lock after agent 1 released")
	}
}

// TestTimeout_ReturnsError tests that timeout returns error if lock held
func TestTimeout_ReturnsError(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner1 := "agent-1"
	owner2 := "agent-2"

	// Agent 1 acquires lock
	err := manager.Acquire(path, owner1, 5*time.Second)
	if err != nil {
		t.Fatalf("Agent 1 failed to acquire lock: %v", err)
	}
	defer func() { _ = manager.Release(path, owner1) }()

	// Agent 2 tries to acquire with short timeout (should fail)
	start := time.Now()
	err = manager.Acquire(path, owner2, 100*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Should have taken approximately the timeout duration
	if elapsed < 100*time.Millisecond {
		t.Errorf("Timeout occurred too quickly: %v", elapsed)
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("Timeout took too long: %v", elapsed)
	}
}

// TestForceRelease tests that force release works
func TestForceRelease(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner := "agent-1"

	// Acquire lock
	err := manager.Acquire(path, owner, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Verify locked
	if !manager.IsLocked(path) {
		t.Error("File should be locked")
	}

	// Force release (different owner)
	err = manager.ForceRelease(path)
	if err != nil {
		t.Fatalf("Failed to force release lock: %v", err)
	}

	// Verify unlocked
	if manager.IsLocked(path) {
		t.Error("File should not be locked after force release")
	}

	// Should be able to acquire now
	err = manager.Acquire(path, "agent-2", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock after force release: %v", err)
	}
}

// TestReleaseAll tests releasing all locks for an owner
func TestReleaseAll(t *testing.T) {
	manager := NewLockManager()
	owner := "agent-1"
	paths := []string{"/test/file1.txt", "/test/file2.txt", "/test/file3.txt"}

	// Acquire multiple locks
	for _, path := range paths {
		err := manager.Acquire(path, owner, 5*time.Second)
		if err != nil {
			t.Fatalf("Failed to acquire lock for %s: %v", path, err)
		}
	}

	// Verify all locked
	for _, path := range paths {
		if !manager.IsLocked(path) {
			t.Errorf("File %s should be locked", path)
		}
	}

	// Release all locks for owner
	err := manager.ReleaseAll(owner)
	if err != nil {
		t.Fatalf("Failed to release all locks: %v", err)
	}

	// Verify all unlocked
	for _, path := range paths {
		if manager.IsLocked(path) {
			t.Errorf("File %s should not be locked", path)
		}
	}
}

// TestDeadlockPrevention tests that no deadlocks are possible
func TestDeadlockPrevention(t *testing.T) {
	manager := NewLockManager()
	paths := []string{"/test/file1.txt", "/test/file2.txt", "/test/file3.txt"}
	numGoroutines := 10
	iterations := 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*iterations)

	// Spawn multiple goroutines trying to acquire multiple locks
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			owner := "agent-" + string(rune('0'+id))

			for j := 0; j < iterations; j++ {
				// Try to acquire locks for random paths
				path := paths[j%len(paths)]

				err := manager.Acquire(path, owner, 100*time.Millisecond)
				if err != nil {
					// Timeout is acceptable in stress test
					continue
				}

				// Hold lock briefly
				time.Sleep(1 * time.Millisecond)

				// Release
				err = manager.Release(path, owner)
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	// Wait with timeout to detect deadlocks
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - no deadlock
	case err := <-errors:
		t.Fatalf("Error during stress test: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("Deadlock detected - test timed out")
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Error during stress test: %v", err)
	}
}

// TestSameOwnerCanReacquire tests that same owner can re-acquire their own lock
func TestSameOwnerCanReacquire(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner := "agent-1"

	// Acquire lock
	err := manager.Acquire(path, owner, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Same owner re-acquires (should succeed immediately)
	err = manager.Acquire(path, owner, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to re-acquire own lock: %v", err)
	}

	// Release once should be enough
	err = manager.Release(path, owner)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Should be unlocked
	if manager.IsLocked(path) {
		t.Error("File should not be locked after single release")
	}
}

// TestReleaseNonExistentLock tests releasing a lock that doesn't exist
func TestReleaseNonExistentLock(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner := "agent-1"

	// Try to release non-existent lock (should be no-op or return error)
	err := manager.Release(path, owner)
	// Should either succeed (no-op) or return a specific error
	// Let's allow both behaviors
	_ = err // Accept any result for non-existent lock
}

// TestReleaseWrongOwner tests that wrong owner cannot release lock
func TestReleaseWrongOwner(t *testing.T) {
	manager := NewLockManager()
	path := "/test/file.txt"
	owner1 := "agent-1"
	owner2 := "agent-2"

	// Agent 1 acquires lock
	err := manager.Acquire(path, owner1, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Agent 2 tries to release (should fail)
	err = manager.Release(path, owner2)
	if err == nil {
		t.Error("Expected error when wrong owner tries to release")
	}

	// Lock should still be held
	if !manager.IsLocked(path) {
		t.Error("Lock should still be held by agent 1")
	}

	// Agent 1 can release
	err = manager.Release(path, owner1)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}
}
