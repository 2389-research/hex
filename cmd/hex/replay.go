// ABOUTME: Replay subcommand for event timeline viewing
// ABOUTME: Provides access to hexreplay tool

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay [event-file]",
	Short: "Replay event timeline from event file",
	Long: `Display chronological timeline of agent events.

Examples:
  hex replay /tmp/hex_events_20251208.jsonl
  hex replay events.jsonl --agent root.1
  hex replay events.jsonl --type ToolCall`,
	Args: cobra.ExactArgs(1),
	RunE: runReplay,
}

var (
	replayAgentFilter string
	replayTypeFilter  string
)

func init() {
	replayCmd.Flags().StringVar(&replayAgentFilter, "agent", "", "filter by agent ID")
	replayCmd.Flags().StringVar(&replayTypeFilter, "type", "", "filter by event type")
}

func runReplay(cmd *cobra.Command, args []string) error {
	eventFile := args[0]

	// Check if event file exists
	if _, err := os.Stat(eventFile); os.IsNotExist(err) {
		return fmt.Errorf("event file not found: %s", eventFile)
	}

	// Find hexreplay binary
	hexreplayPath, err := exec.LookPath("hexreplay")
	if err != nil {
		// Try in bin/ directory
		binPath := filepath.Join("bin", "hexreplay")
		if _, err := os.Stat(binPath); err == nil {
			hexreplayPath = binPath
		} else {
			return fmt.Errorf("hexreplay not found. Run: go build -o bin/hexreplay ./cmd/hexreplay")
		}
	}

	// Build command arguments
	cmdArgs := []string{"-events", eventFile}

	if replayAgentFilter != "" {
		cmdArgs = append(cmdArgs, "-agent", replayAgentFilter)
	}

	if replayTypeFilter != "" {
		cmdArgs = append(cmdArgs, "-type", replayTypeFilter)
	}

	// Execute hexreplay
	replayExec := exec.Command(hexreplayPath, cmdArgs...)
	replayExec.Stdout = os.Stdout
	replayExec.Stderr = os.Stderr

	return replayExec.Run()
}
