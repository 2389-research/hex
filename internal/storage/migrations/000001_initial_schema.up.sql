-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT 'New Conversation',
    model TEXT NOT NULL,
    system_prompt TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_favorite BOOLEAN DEFAULT 0
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tool_calls JSON,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- Todos table
CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL CHECK(length(trim(content)) > 0),
    active_form TEXT NOT NULL CHECK(length(trim(active_form)) > 0),
    status TEXT NOT NULL CHECK(status IN ('pending', 'in_progress', 'completed')),
    conversation_id TEXT REFERENCES conversations(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- History table
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

-- Indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation
    ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_conversations_updated
    ON conversations(updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_conversations_favorite
    ON conversations(is_favorite) WHERE is_favorite = 1;

CREATE INDEX IF NOT EXISTS idx_todos_status
    ON todos(status);

CREATE INDEX IF NOT EXISTS idx_todos_conversation
    ON todos(conversation_id);

CREATE INDEX IF NOT EXISTS idx_todos_created
    ON todos(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_history_conversation
    ON history(conversation_id);

CREATE INDEX IF NOT EXISTS idx_history_created
    ON history(created_at DESC);
