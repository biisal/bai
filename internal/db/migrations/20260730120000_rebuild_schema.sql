-- +goose Up
-- Rebuild schema: fix SERIAL → INTEGER PRIMARY KEY AUTOINCREMENT for SQLite compat,
-- drop unused users table, add is_error for Anthropic tool support.

DROP TABLE IF EXISTS tool_call;
DROP TABLE IF EXISTS chat_completion;
DROP TABLE IF EXISTS users;

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
    content TEXT NOT NULL DEFAULT '',
    error TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tool_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    call_id TEXT NOT NULL,
    name TEXT NOT NULL,
    arguments TEXT NOT NULL,           -- JSON string (OpenAI convention)
    result TEXT,                       -- NULL until executed
    is_error INTEGER NOT NULL DEFAULT 0,  -- Anthropic tool_result.is_error
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'success', 'error')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS tool_calls;
DROP TABLE IF EXISTS messages;
