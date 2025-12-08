-- Add provider column to conversations table
ALTER TABLE conversations ADD COLUMN provider TEXT DEFAULT 'anthropic';
