// ABOUTME: Commands for managing conversation favorites
// ABOUTME: Toggle favorite status and list all favorite conversations
package main

import (
	"fmt"
	"time"

	"github.com/2389-research/hex/internal/storage"
	"github.com/spf13/cobra"
)

var favoriteCmd = &cobra.Command{
	Use:   "favorite <conversation-id>",
	Short: "Toggle favorite status of a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		convID := args[0]

		// Open database
		db, err := openDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer func() { _ = db.Close() }()

		// Get current conversation to check if it exists and get current favorite status
		conv, err := storage.GetConversation(db, convID)
		if err != nil {
			return fmt.Errorf("conversation not found: %w", err)
		}

		// Toggle favorite status
		newStatus := !conv.IsFavorite
		if err := storage.SetFavorite(db, convID, newStatus); err != nil {
			return fmt.Errorf("set favorite: %w", err)
		}

		// Print confirmation
		if newStatus {
			fmt.Printf("⭐ Marked '%s' as favorite\n", conv.Title)
		} else {
			fmt.Printf("Removed '%s' from favorites\n", conv.Title)
		}

		return nil
	},
}

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "List all favorite conversations",
	RunE: func(_ *cobra.Command, _ []string) error {
		// Open database
		db, err := openDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer func() { _ = db.Close() }()

		// Get all favorites
		favorites, err := storage.ListFavorites(db)
		if err != nil {
			return fmt.Errorf("list favorites: %w", err)
		}

		// Print results
		if len(favorites) == 0 {
			fmt.Println("No favorite conversations yet.")
			fmt.Println("Use 'hex favorite <conv-id>' to mark a conversation as favorite.")
			return nil
		}

		fmt.Printf("⭐ Favorite Conversations (%d)\n\n", len(favorites))
		for _, conv := range favorites {
			// Format timestamp
			relTime := formatRelativeTime(conv.UpdatedAt)

			// Truncate title if too long
			title := conv.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}

			fmt.Printf("  ⭐ %s\n", title)
			fmt.Printf("     ID: %s | Model: %s | Updated: %s\n\n",
				conv.ID, conv.Model, relTime)
		}

		return nil
	},
}

// formatRelativeTime formats a time as relative to now (e.g., "2 hours ago")
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		if months < 12 {
			return fmt.Sprintf("%d months ago", months)
		}
		years := months / 12
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func init() {
	rootCmd.AddCommand(favoriteCmd)
	rootCmd.AddCommand(favoritesCmd)
}
