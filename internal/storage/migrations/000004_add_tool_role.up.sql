-- Add 'tool' to allowed message roles
-- SQLite doesn't support ALTER CHECK, so we need to recreate the table

-- Create new table with updated constraint (includes all columns from migrations 1-2)
CREATE TABLE messages_new (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    tool_calls JSON,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_summary BOOLEAN DEFAULT 0,
    provider TEXT,
    model TEXT
);

-- Copy data from old table
INSERT INTO messages_new SELECT * FROM messages;

-- Drop old table
DROP TABLE messages;

-- Rename new table
ALTER TABLE messages_new RENAME TO messages;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation
    ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_messages_is_summary
    ON messages(conversation_id, is_summary)
    WHERE is_summary = 1;
