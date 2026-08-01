-- +goose Up
CREATE TABLE tool_call (
	id SERIAL PRIMARY KEY,
	call_id TEXT,
	name TEXT,
	arguments TEXT,
	result TEXT DEFAULT NULL,
	status TEXT NOT NULL DEFAULT 'success'
	CHECK(status IN ('pending','success','error')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completion_id INTEGER NOT NULL
	REFERENCES chat_completion(id)
   	ON DELETE CASCADE
);

-- +goose Down
DROP TABLE tool_call;
