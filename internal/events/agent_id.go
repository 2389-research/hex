// ABOUTME: agent_id.go provides hierarchical agent ID generation for tracking agent trees.
// ABOUTME: IDs follow the pattern: root, root.1, root.2, root.1.1, etc.

package events

import (
	"fmt"
	"sync"
)

var (
	childCounters = make(map[string]int)
	counterMu     sync.Mutex
)

// GenerateAgentID generates a hierarchical agent ID based on the parent ID.
// Returns "root" for empty parent, or "parent.N" for children.
func GenerateAgentID(parentID string) string {
	if parentID == "" {
		return "root"
	}

	counterMu.Lock()
	defer counterMu.Unlock()

	childCounters[parentID]++
	return fmt.Sprintf("%s.%d", parentID, childCounters[parentID])
}

// ResetCounters resets the child counters (useful for testing)
func ResetCounters() {
	counterMu.Lock()
	defer counterMu.Unlock()
	childCounters = make(map[string]int)
}
