-- ABOUTME: Todo list persistence schema
-- ABOUTME: Stores todo items with status tracking and optional conversation association

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

CREATE INDEX IF NOT EXISTS idx_todos_status
    ON todos(status);

CREATE INDEX IF NOT EXISTS idx_todos_conversation
    ON todos(conversation_id);

CREATE INDEX IF NOT EXISTS idx_todos_created
    ON todos(created_at DESC);
