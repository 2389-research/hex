-- ABOUTME: Add favorites support to conversations table
-- ABOUTME: Adds is_favorite column and index for filtering

-- SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN
-- This migration will fail if run twice on the same database
-- TODO: Implement proper migration tracking to prevent duplicate runs
-- For now, we handle the error gracefully in code

-- Add is_favorite column (will error if column already exists - that's OK)
-- The schema initialization code should ignore this specific error
ALTER TABLE conversations ADD COLUMN is_favorite BOOLEAN DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_conversations_favorite
    ON conversations(is_favorite) WHERE is_favorite = 1;
