-- Remove 'tool' role - recreate table with original constraint
-- Note: This will fail if there are any messages with role='tool'

CREATE TABLE messages_new (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tool_calls JSON,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_summary BOOLEAN DEFAULT 0,
    provider TEXT,
    model TEXT,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- Copy data (will fail if tool messages exist)
INSERT INTO messages_new SELECT * FROM messages;

DROP TABLE messages;

ALTER TABLE messages_new RENAME TO messages;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation
    ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_messages_is_summary
    ON messages(conversation_id, is_summary)
    WHERE is_summary = 1;
