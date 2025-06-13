-- Create message_trees table
CREATE TABLE message_trees (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ai_id INTEGER, -- Will reference models table when implemented
    title VARCHAR(255) DEFAULT 'New Chat',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create messages table
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    message_tree_id INTEGER NOT NULL REFERENCES message_trees(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_incoming BOOLEAN NOT NULL DEFAULT FALSE, -- false = user message, true = AI response
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_message_trees_user_id ON message_trees(user_id);
CREATE INDEX idx_messages_tree_id ON messages(message_tree_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);