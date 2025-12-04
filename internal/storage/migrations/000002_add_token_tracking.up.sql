-- Add token tracking to conversations
ALTER TABLE conversations ADD COLUMN prompt_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN completion_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN total_cost REAL DEFAULT 0.0;

-- Add summary tracking
ALTER TABLE messages ADD COLUMN is_summary BOOLEAN DEFAULT 0;
ALTER TABLE conversations ADD COLUMN summary_message_id INTEGER REFERENCES messages(id);

-- Add provider tracking to messages
ALTER TABLE messages ADD COLUMN provider TEXT;
ALTER TABLE messages ADD COLUMN model TEXT;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_messages_is_summary
    ON messages(conversation_id, is_summary)
    WHERE is_summary = 1;
