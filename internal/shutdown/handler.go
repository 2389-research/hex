// ABOUTME: Shutdown handler for graceful cascading shutdown on SIGINT/SIGTERM
// ABOUTME: Listens for OS signals and triggers process registry cleanup

package shutdown

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/2389-research/hex/internal/filelock"
	"github.com/2389-research/hex/internal/registry"
)

// InitShutdownHandler sets up signal handlers for graceful shutdown
// When SIGINT or SIGTERM is received, all child processes are stopped
func InitShutdownHandler() {
	InitShutdownHandlerWithRegistry(registry.Global(), nil)
}

// InitShutdownHandlerWithRegistry is the testable version that accepts a registry
// stopChan is optional and used for testing to signal when shutdown is complete
func InitShutdownHandlerWithRegistry(reg *registry.ProcessRegistry, stopChan chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down gracefully...\n", sig)

		// Get the current agent ID from environment
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID == "" {
			agentID = "root" // Default to root if not set
		}

		// Release all file locks held by this agent
		lockManager := filelock.Global()
		if err := lockManager.ReleaseAll(agentID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error releasing file locks: %v\n", err)
		}

		// Stop all child processes recursively
		if err := reg.StopCascading(agentID, func(id string) {
			fmt.Fprintf(os.Stderr, "Stopped process: %s\n", id)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error during cascading shutdown: %v\n", err)
		}

		// Check for orphans
		orphans := reg.GetOrphans()
		if len(orphans) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: found orphaned processes: %v\n", orphans)
			// Try to stop orphans
			for _, orphanID := range orphans {
				if err := reg.StopCascading(orphanID, nil); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to stop orphan %s: %v\n", orphanID, err)
				}
			}
		}

		// Signal test that we're done
		if stopChan != nil {
			close(stopChan)
			// In test mode, don't exit
			return
		}

		// Exit the process
		os.Exit(0)
	}()
}
