// ABOUTME: Todo repository for CRUD operations on todo items
// ABOUTME: Handles saving, loading, and clearing todo lists with optional conversation scoping
package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Todo represents a single todo item
type Todo struct {
	ID             string
	Content        string
	ActiveForm     string
	Status         string // "pending", "in_progress", "completed"
	ConversationID *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SaveTodos replaces all todos for the given conversation (or global if nil) with the provided list
func SaveTodos(db *sql.DB, todos []Todo, conversationID *string) error {
	// Start transaction for atomic replace operation
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing todos for this conversation (or all global if nil)
	var deleteQuery string
	var deleteArgs []interface{}

	if conversationID != nil {
		deleteQuery = `DELETE FROM todos WHERE conversation_id = ?`
		deleteArgs = []interface{}{*conversationID}
	} else {
		deleteQuery = `DELETE FROM todos WHERE conversation_id IS NULL`
		deleteArgs = []interface{}{}
	}

	if _, err := tx.Exec(deleteQuery, deleteArgs...); err != nil {
		return fmt.Errorf("delete existing todos: %w", err)
	}

	// Insert new todos
	insertQuery := `
		INSERT INTO todos (id, content, active_form, status, conversation_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	for i, todo := range todos {
		// Generate ID if not present
		id := todo.ID
		if id == "" {
			id = uuid.New().String()
		}

		// For timestamps: preserve existing values or default to now
		// When a todo has an ID and created_at, it's being updated
		// When a todo has no ID, it's new
		createdAt := todo.CreatedAt
		if createdAt.IsZero() {
			createdAt = now
		}

		// Always update the updated_at timestamp to now for saves
		updatedAt := now

		// Use conversation ID from parameter if todo doesn't have one
		todoConvID := todo.ConversationID
		if todoConvID == nil {
			todoConvID = conversationID
		}

		_, err := tx.Exec(insertQuery, id, todo.Content, todo.ActiveForm, todo.Status, todoConvID, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert todo at index %d: %w", i, err)
		}

		// Update the todo with generated values for caller
		todos[i].ID = id
		todos[i].CreatedAt = createdAt
		todos[i].UpdatedAt = updatedAt
		if todos[i].ConversationID == nil {
			todos[i].ConversationID = todoConvID
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// LoadTodos retrieves all todos for the given conversation (or global if nil), ordered by created_at ASC
func LoadTodos(db *sql.DB, conversationID *string) ([]Todo, error) {
	var query string
	var args []interface{}

	if conversationID != nil {
		query = `
			SELECT id, content, active_form, status, conversation_id, created_at, updated_at
			FROM todos
			WHERE conversation_id = ?
			ORDER BY created_at ASC
		`
		args = []interface{}{*conversationID}
	} else {
		query = `
			SELECT id, content, active_form, status, conversation_id, created_at, updated_at
			FROM todos
			WHERE conversation_id IS NULL
			ORDER BY created_at ASC
		`
		args = []interface{}{}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var convID sql.NullString

		err := rows.Scan(
			&todo.ID,
			&todo.Content,
			&todo.ActiveForm,
			&todo.Status,
			&convID,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}

		if convID.Valid {
			todo.ConversationID = &convID.String
		}

		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return todos, nil
}

// ClearCompleted removes all completed todos for the given conversation (or global if nil)
func ClearCompleted(db *sql.DB, conversationID *string) error {
	var query string
	var args []interface{}

	if conversationID != nil {
		query = `DELETE FROM todos WHERE status = 'completed' AND conversation_id = ?`
		args = []interface{}{*conversationID}
	} else {
		query = `DELETE FROM todos WHERE status = 'completed' AND conversation_id IS NULL`
		args = []interface{}{}
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("clear completed todos: %w", err)
	}

	return nil
}
