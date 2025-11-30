// ABOUTME: Performance benchmarks for SQLite storage operations
// ABOUTME: Measures conversation, message CRUD, and search performance
package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// setupBenchDB creates a temporary database for benchmarking
func setupBenchDB(b *testing.B) *sql.DB {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	db, err := OpenDatabase(dbPath)
	if err != nil {
		b.Fatalf("failed to open database: %v", err)
	}

	return db
}

// BenchmarkConversationCreate measures conversation creation
func BenchmarkConversationCreate(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv := &Conversation{
			Title: fmt.Sprintf("Conversation %d", i),
		}
		if err := CreateConversation(db, conv); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageInsert measures single message insertion
func BenchmarkMessageInsert(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	// Create a conversation first
	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        fmt.Sprintf("Message %d", i),
		}
		if err := CreateMessage(db, msg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageGet measures message retrieval by ID
func BenchmarkMessageGet(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	// Create some messages
	messageIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		msg := &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        fmt.Sprintf("Message %d", i),
		}
		if err := CreateMessage(db, msg); err != nil {
			b.Fatal(err)
		}
		messageIDs[i] = msg.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := messageIDs[i%len(messageIDs)]
		_, err := GetMessage(db, id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConversationList measures listing conversations
func BenchmarkConversationList(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	// Create 100 conversations
	for i := 0; i < 100; i++ {
		conv := &Conversation{
			Title: fmt.Sprintf("Conversation %d", i),
		}
		if err := CreateConversation(db, conv); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ListConversations(db, 20, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTransactionOverhead measures transaction cost
func BenchmarkTransactionOverhead(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, err := db.Begin()
		if err != nil {
			b.Fatal(err)
		}

		msg := &Message{
			ID:             uuid.New().String(),
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "Test",
			CreatedAt:      time.Now(),
		}

		query := `INSERT INTO messages (id, conversation_id, role, content, tool_calls, metadata, created_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.ToolCalls, msg.Metadata, msg.CreatedAt)
		if err != nil {
			b.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPreparedStatement measures prepared statement performance
func BenchmarkPreparedStatement(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	// Prepare statement once
	stmt, err := db.Prepare(`INSERT INTO messages (id, conversation_id, role, content, created_at)
							 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stmt.Exec(
			uuid.New().String(),
			conv.ID,
			"user",
			fmt.Sprintf("Message %d", i),
			time.Now(),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeMessageContent measures performance with large message content
func BenchmarkLargeMessageContent(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	// Create 1MB of content
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        string(largeContent),
		}
		if err := CreateMessage(db, msg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentReads measures parallel read performance
func BenchmarkConcurrentReads(b *testing.B) {
	db := setupBenchDB(b)
	defer func() { _ = db.Close() }()

	// Create test data
	conv := &Conversation{Title: "Test Conversation"}
	if err := CreateConversation(db, conv); err != nil {
		b.Fatal(err)
	}

	messageIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		msg := &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        fmt.Sprintf("Message %d", i),
		}
		if err := CreateMessage(db, msg); err != nil {
			b.Fatal(err)
		}
		messageIDs[i] = msg.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := messageIDs[i%len(messageIDs)]
			_, err := GetMessage(db, id)
			if err != nil {
				b.Error(err)
			}
			i++
		}
	})
}
