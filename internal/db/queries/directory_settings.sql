-- name: GetDirectorySettings :one
SELECT * FROM directory_settings
WHERE directory = ?1;

-- name: UpsertDirectorySettings :exec
INSERT INTO directory_settings (directory, auto_git_init)
VALUES (?1, ?2)
ON CONFLICT(directory) DO UPDATE SET auto_git_init = excluded.auto_git_init;
