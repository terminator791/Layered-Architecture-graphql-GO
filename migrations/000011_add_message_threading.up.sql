-- Add threading support to messages table

-- Add new columns for threading
ALTER TABLE messages 
ADD COLUMN thread_id VARCHAR(255),
ADD COLUMN is_thread_root BOOLEAN DEFAULT FALSE;

-- Create index for efficient thread queries
CREATE INDEX idx_messages_thread_id ON messages(thread_id, created_at ASC) 
WHERE thread_id IS NOT NULL AND deleted_at IS NULL;

-- Create index for finding thread roots
CREATE INDEX idx_messages_thread_root ON messages(is_thread_root, room_id, created_at DESC) 
WHERE is_thread_root = TRUE AND deleted_at IS NULL;

-- Add foreign key constraint for thread_id referencing the root message
ALTER TABLE messages 
ADD CONSTRAINT fk_messages_thread_id 
FOREIGN KEY (thread_id) REFERENCES messages(id) 
DEFERRABLE INITIALLY DEFERRED;

-- Add check constraint to ensure thread roots don't have thread_id
ALTER TABLE messages 
ADD CONSTRAINT chk_thread_root_consistency 
CHECK (
    (is_thread_root = TRUE AND thread_id IS NULL) OR 
    (is_thread_root = FALSE)
);