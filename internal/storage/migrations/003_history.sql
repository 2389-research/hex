-- ABOUTME: History table with FTS5 for searchable command history
-- ABOUTME: Stores all user messages and assistant responses with full-text search

-- Main history table
CREATE TABLE IF NOT EXISTS history (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_message TEXT NOT NULL,
    assistant_response TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS history_fts USING fts5(
    user_message,
    assistant_response,
    content=history,
    content_rowid=rowid
);

-- Triggers to keep FTS5 in sync
CREATE TRIGGER IF NOT EXISTS history_fts_insert AFTER INSERT ON history BEGIN
    INSERT INTO history_fts(rowid, user_message, assistant_response)
    VALUES (new.rowid, new.user_message, new.assistant_response);
END;

CREATE TRIGGER IF NOT EXISTS history_fts_delete AFTER DELETE ON history BEGIN
    DELETE FROM history_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS history_fts_update AFTER UPDATE ON history BEGIN
    DELETE FROM history_fts WHERE rowid = old.rowid;
    INSERT INTO history_fts(rowid, user_message, assistant_response)
    VALUES (new.rowid, new.user_message, new.assistant_response);
END;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_history_conversation
    ON history(conversation_id);

CREATE INDEX IF NOT EXISTS idx_history_created
    ON history(created_at DESC);
