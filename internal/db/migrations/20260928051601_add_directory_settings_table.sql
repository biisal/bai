-- +goose Up
CREATE TABLE directory_settings (
    id SERIAL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    auto_git_init BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE directory_settings;