-- Create messages table for storing chat messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    room VARCHAR(50) NOT NULL,
    "user" VARCHAR(50) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_room_created_at ON messages(room, created_at);