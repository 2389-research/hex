DROP INDEX IF EXISTS idx_messages_is_summary;

-- SQLite doesn't support DROP COLUMN, so we'd need to recreate tables
-- For now, down migration is not supported for this change
-- In production, we'd implement full table recreation
