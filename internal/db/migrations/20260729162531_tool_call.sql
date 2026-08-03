-- +goose Up
CREATE TABLE tool_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    call_id TEXT NOT NULL,
    name TEXT NOT NULL,
    arguments TEXT NOT NULL,           
    result TEXT,                       
    is_error INTEGER NOT NULL DEFAULT 0,  
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'success', 'error')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE tool_calls;
