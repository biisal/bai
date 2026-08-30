-- +goose Up
CREATE TABLE user_settings (
	id SERIAL PRIMARY KEY CHECK (id = 1),
    theme VARCHAR(255)
);

-- +goose Down
DROP TABLE user_settings;