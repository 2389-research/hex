// ABOUTME: File lock manager that prevents race conditions when multiple agents write to the same file
// ABOUTME: Provides thread-safe lock acquisition, release, and timeout mechanisms

package filelock

import (
	"fmt"
	"sync"
	"time"
)

// FileLock represents a lock on a file
type FileLock struct {
	Path     string
	OwnerID  string
	LockedAt time.Time
}

// LockManager manages file locks across multiple agents
type LockManager struct {
	mu    sync.Mutex
	locks map[string]*FileLock
}

var (
	globalManager     *LockManager
	globalManagerOnce sync.Once
)

// NewLockManager creates a new lock manager
func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*FileLock),
	}
}

// Global returns the singleton lock manager
func Global() *LockManager {
	globalManagerOnce.Do(func() {
		globalManager = NewLockManager()
	})
	return globalManager
}

// Acquire attempts to acquire a lock on a file
// Blocks until lock is available or timeout is reached
func (m *LockManager) Acquire(path, ownerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		m.mu.Lock()

		// Check if locked
		lock, exists := m.locks[path]
		if !exists || lock.OwnerID == ownerID {
			// Not locked or we already own it - acquire it
			m.locks[path] = &FileLock{
				Path:     path,
				OwnerID:  ownerID,
				LockedAt: time.Now(),
			}
			m.mu.Unlock()
			return nil
		}

		// Lock is held by someone else
		currentOwner := lock.OwnerID
		m.mu.Unlock()

		// Check timeout
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout acquiring lock for %s (held by %s)", path, currentOwner)
		}

		// Wait a bit before retrying
		time.Sleep(50 * time.Millisecond)
	}
}

// Release releases a lock on a file
// Returns error if lock is not held by the specified owner
func (m *LockManager) Release(path, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[path]
	if !exists {
		// Lock doesn't exist - this is a no-op
		return nil
	}

	if lock.OwnerID != ownerID {
		return fmt.Errorf("cannot release lock for %s: owned by %s, not %s", path, lock.OwnerID, ownerID)
	}

	delete(m.locks, path)
	return nil
}

// ForceRelease forcibly releases a lock on a file
// Used for cleanup when an agent crashes
func (m *LockManager) ForceRelease(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, path)
	return nil
}

// IsLocked checks if a file is currently locked
func (m *LockManager) IsLocked(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.locks[path]
	return exists
}

// ReleaseAll releases all locks held by a specific owner
// Used when an agent shuts down
func (m *LockManager) ReleaseAll(ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and delete all locks for this owner
	for path, lock := range m.locks {
		if lock.OwnerID == ownerID {
			delete(m.locks, path)
		}
	}

	return nil
}
