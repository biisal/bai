-- +goose Up
CREATE TABLE chat_completion (
	id SERIAL PRIMARY KEY,
	conversation_id INTEGER REFERENCES conversations (id),
	role TEXT check (role in ('user', 'assistant' )),
	content TEXT,
	error TEXT
);

-- +goose Down
DROP TABLE chat_completion;
