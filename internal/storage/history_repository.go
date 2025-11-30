// ABOUTME: Repository for command history with FTS5 full-text search
// ABOUTME: Manages storage and retrieval of user messages and assistant responses
package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// HistoryEntry represents a single command history entry
type HistoryEntry struct {
	ID                string
	ConversationID    string
	UserMessage       string
	AssistantResponse string
	CreatedAt         time.Time
}

// AddHistoryEntry saves a history entry to the database
func AddHistoryEntry(db *sql.DB, entry *HistoryEntry) error {
	query := `
		INSERT INTO history (id, conversation_id, user_message, assistant_response, created_at)
		VALUES (?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))
	`

	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := db.Exec(query, entry.ID, entry.ConversationID, entry.UserMessage, entry.AssistantResponse, createdAt)
	if err != nil {
		return fmt.Errorf("insert history entry: %w", err)
	}

	return nil
}

// SearchHistory performs FTS5 search on history
func SearchHistory(db *sql.DB, query string, limit int) ([]*HistoryEntry, error) {
	sqlQuery := `
		SELECT h.id, h.conversation_id, h.user_message, h.assistant_response, h.created_at
		FROM history h
		JOIN history_fts ON history_fts.rowid = h.rowid
		WHERE history_fts MATCH ?
		ORDER BY h.created_at DESC
		LIMIT ?
	`

	rows, err := db.Query(sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*HistoryEntry
	for rows.Next() {
		entry := &HistoryEntry{}
		err := rows.Scan(&entry.ID, &entry.ConversationID, &entry.UserMessage, &entry.AssistantResponse, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan history entry: %w", err)
		}
		results = append(results, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}

	return results, nil
}

// GetRecentHistory retrieves the most recent history entries
func GetRecentHistory(db *sql.DB, limit int) ([]*HistoryEntry, error) {
	query := `
		SELECT id, conversation_id, user_message, assistant_response, created_at
		FROM history
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*HistoryEntry
	for rows.Next() {
		entry := &HistoryEntry{}
		err := rows.Scan(&entry.ID, &entry.ConversationID, &entry.UserMessage, &entry.AssistantResponse, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan history entry: %w", err)
		}
		results = append(results, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}

	return results, nil
}
