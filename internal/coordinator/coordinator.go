// ABOUTME: Global resource coordinator for multi-agent file locking
// ABOUTME: Wraps mux coordinator with singleton pattern for hex-wide access

package coordinator

import (
	"context"
	"sync"

	muxcoord "github.com/2389-research/mux/coordinator"
)

var (
	global     *muxcoord.Coordinator
	globalOnce sync.Once
)

// Global returns the singleton coordinator instance.
// Thread-safe via sync.Once.
func Global() *muxcoord.Coordinator {
	globalOnce.Do(func() {
		global = muxcoord.New()
	})
	return global
}

// Acquire acquires a lock on a resource for the given agent.
// Convenience wrapper around Global().Acquire().
func Acquire(ctx context.Context, agentID, resourceID string) error {
	return Global().Acquire(ctx, agentID, resourceID)
}

// Release releases a lock on a resource.
// Convenience wrapper around Global().Release().
func Release(agentID, resourceID string) error {
	return Global().Release(agentID, resourceID)
}

// ReleaseAll releases all locks held by the given agent.
// Convenience wrapper around Global().ReleaseAll().
func ReleaseAll(agentID string) {
	Global().ReleaseAll(agentID)
}
