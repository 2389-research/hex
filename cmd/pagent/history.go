// ABOUTME: History command for viewing and searching command history
// ABOUTME: Provides recent history listing and FTS5 search capabilities

package main

import (
	"fmt"
	"strings"

	"github.com/harper/pagent/internal/storage"
	"github.com/spf13/cobra"
)

var (
	historyLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View command history",
	Long: `View your command history with Pagen.

Shows recent conversations with timestamps, message previews, and conversation IDs.
Use the search subcommand to find specific topics.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runHistory(cmd)
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search command history",
	Long: `Search your command history using full-text search.

Searches both user messages and assistant responses for the given query.
Results are ranked by relevance and sorted by recency.

Examples:
  clem history search "docker"
  clem history search "bug fix"
  clem history search "python script"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistorySearch(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historySearchCmd)

	// Add --limit flag to both commands
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "Number of results to show")
	historySearchCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "Number of results to show")
}

func runHistory(cmd *cobra.Command) error {
	// Open database
	db, err := openDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Get recent history
	entries, err := storage.GetRecentHistory(db, historyLimit)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(entries) == 0 {
		cmd.Println("No history found.")
		cmd.Println()
		cmd.Println("Start a conversation with Pagen to build your history!")
		return nil
	}

	// Display results
	cmd.Printf("Recent history (showing %d):\n", len(entries))
	cmd.Println()

	for _, entry := range entries {
		displayHistoryEntry(cmd, entry)
	}

	cmd.Printf("\nTotal: %d entr", len(entries))
	if len(entries) == 1 {
		cmd.Println("y")
	} else {
		cmd.Println("ies")
	}
	cmd.Printf("Use 'clem history --limit N' to show more results\n")
	cmd.Printf("Use 'clem history search \"query\"' to search\n")

	return nil
}

func runHistorySearch(cmd *cobra.Command, query string) error {
	// Open database
	db, err := openDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Search history
	entries, err := storage.SearchHistory(db, query, historyLimit)
	if err != nil {
		return fmt.Errorf("failed to search history: %w", err)
	}

	if len(entries) == 0 {
		cmd.Printf("No results found for: %q\n", query)
		cmd.Println()
		cmd.Println("Try different search terms or check your spelling.")
		return nil
	}

	// Display results
	cmd.Printf("Search results for %q (showing %d):\n", query, len(entries))
	cmd.Println()

	for _, entry := range entries {
		displayHistoryEntry(cmd, entry)
	}

	cmd.Printf("\nTotal: %d result", len(entries))
	if len(entries) == 1 {
		cmd.Println()
	} else {
		cmd.Println("s")
	}
	cmd.Printf("Use 'clem history search \"query\" --limit N' to show more results\n")

	return nil
}

// displayHistoryEntry formats and displays a single history entry
func displayHistoryEntry(cmd *cobra.Command, entry *storage.HistoryEntry) {
	// Format timestamp as relative time
	relTime := formatRelativeTime(entry.CreatedAt)

	// Truncate message to 60 chars
	preview := truncateString(entry.UserMessage, 60)

	// Display formatted entry
	cmd.Printf("  %s - %s\n", relTime, preview)
	cmd.Printf("    Conversation: %s\n", entry.ConversationID)
	cmd.Println()
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	// Normalize whitespace (replace newlines/tabs with spaces, collapse multiple spaces)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")

	if len(s) <= maxLen {
		return s
	}

	// Find last space before maxLen-3 to avoid cutting words
	truncateAt := maxLen - 3
	lastSpace := strings.LastIndex(s[:truncateAt], " ")
	if lastSpace > 0 && lastSpace > maxLen/2 {
		// Use last space if it's not too far back
		return s[:lastSpace] + "..."
	}

	// Otherwise just truncate at maxLen-3
	return s[:truncateAt] + "..."
}
