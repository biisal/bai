-- name: CreateConversation :one
INSERT INTO conversations (title , directory) VALUES (?1 , ?2) RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = ?1;

-- name: ListConversations :many
SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC;

-- name: UpdateConversation :exec
UPDATE conversations SET title = ?1, updated_at = datetime('now') WHERE id = ?2;

-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = ?1;

-- name: GetConversationMessageCount :one
SELECT COUNT(*) FROM messages WHERE conversation_id = ?1;

-- name: GetConversationsByDirectory :many
SELECT * FROM conversations WHERE directory = ?1;
