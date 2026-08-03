-- +goose Up
CREATE TABLE provider (
	id SERIAL PRIMARY KEY CHECK (id = 1),
	name TEXT NOT NULL,
	provider_id TEXT UNIQUE,
	model_id TEXT UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE provider;
