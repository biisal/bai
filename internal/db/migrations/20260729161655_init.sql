-- +goose Up
CREATE TABLE chat_completion (
	id SERIAL PRIMARY KEY,
	conversation_id INTEGER REFERENCES conversations (id),
	role TEXT check (role in ('user', 'assistant' )),
	content TEXT,
	error TEXT
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL
        REFERENCES conversations(id) ON DELETE CASCADE,

    role TEXT NOT NULL
        CHECK(role IN ('system', 'user', 'assistant', 'tool')),

    parts TEXT NOT NULL DEFAULT '[]',

    error TEXT,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME
);


CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
	directory TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE provider (
	id SERIAL PRIMARY KEY CHECK (id = 1),
	provider_name TEXT UNIQUE,
	model_id TEXT UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE chat_completion;
DROP TABLE messages;
DROP TABLE conversations;
DROP TABLE provider;
