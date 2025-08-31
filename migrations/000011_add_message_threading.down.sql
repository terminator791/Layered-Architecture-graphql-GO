-- Reverse threading support migration

-- Remove check constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS chk_thread_root_consistency;

-- Remove foreign key constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_thread_id;

-- Remove indexes
DROP INDEX IF EXISTS idx_messages_thread_root;
DROP INDEX IF EXISTS idx_messages_thread_id;

-- Remove columns
ALTER TABLE messages 
DROP COLUMN IF EXISTS is_thread_root,
DROP COLUMN IF EXISTS thread_id;