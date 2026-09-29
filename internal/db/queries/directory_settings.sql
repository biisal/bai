-- name: GetDirectorySettings :one
SELECT * FROM directory_settings
WHERE directory = ?1;

-- name: UpdateDirectorySettings :exec
UPDATE directory_settings
SET auto_git_init = ?2
WHERE directory = ?1;

-- name: CreateDirectorySettings :exec
INSERT INTO directory_settings (directory, auto_git_init)
VALUES (?1, ?2);