-- name: CreateConversation :one
INSERT INTO conversations (title , directory) VALUES (?1 , ?2) RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = ?1;

-- name: GetConversationsByDirectory :many
SELECT * FROM conversations WHERE directory = ?1 ORDER by id DESC;