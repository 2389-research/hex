// ABOUTME: Visualization subcommand for multi-agent execution analysis
// ABOUTME: Provides access to hexviz and hexreplay tools

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize [event-file]",
	Short: "Visualize multi-agent execution from event file",
	Long: `Visualize multi-agent execution history using the hexviz tool.

By default shows tree view. Use --view flag to select:
  tree     - Hierarchical agent structure (default)
  timeline - Chronological event sequence
  cost     - Cost breakdown by agent

Examples:
  hex visualize /tmp/hex_events_20251208.jsonl
  hex visualize events.jsonl --view timeline
  hex visualize events.jsonl --view cost --agent root.1`,
	Args: cobra.ExactArgs(1),
	RunE: runVisualize,
}

var (
	viewMode    string
	agentFilter string
	typeFilter  string
	htmlOutput  string
)

func init() {
	visualizeCmd.Flags().StringVar(&viewMode, "view", "tree", "view mode: tree, timeline, or cost")
	visualizeCmd.Flags().StringVar(&agentFilter, "agent", "", "filter by agent ID")
	visualizeCmd.Flags().StringVar(&typeFilter, "type", "", "filter by event type")
	visualizeCmd.Flags().StringVar(&htmlOutput, "html", "", "export to HTML file")
}

func runVisualize(cmd *cobra.Command, args []string) error {
	eventFile := args[0]

	// Check if event file exists
	if _, err := os.Stat(eventFile); os.IsNotExist(err) {
		return fmt.Errorf("event file not found: %s", eventFile)
	}

	// Find hexviz binary
	hexvizPath, err := exec.LookPath("hexviz")
	if err != nil {
		// Try in bin/ directory
		binPath := filepath.Join("bin", "hexviz")
		if _, err := os.Stat(binPath); err == nil {
			hexvizPath = binPath
		} else {
			return fmt.Errorf("hexviz not found. Run: go build -o bin/hexviz ./cmd/hexviz")
		}
	}

	// Build command arguments
	cmdArgs := []string{
		"-events", eventFile,
		"-view", viewMode,
	}

	if agentFilter != "" {
		cmdArgs = append(cmdArgs, "-agent", agentFilter)
	}

	if typeFilter != "" {
		cmdArgs = append(cmdArgs, "-type", typeFilter)
	}

	if htmlOutput != "" {
		cmdArgs = append(cmdArgs, "-html", htmlOutput)
	}

	// Execute hexviz
	vizCmd := exec.Command(hexvizPath, cmdArgs...)
	vizCmd.Stdout = os.Stdout
	vizCmd.Stderr = os.Stderr

	return vizCmd.Run()
}
