-- name: CreateConversation :one
INSERT INTO conversations (title , directory) VALUES (?1 , ?2) RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = ?1;

-- name: GetConversationsByDirectory :many
SELECT * FROM conversations WHERE directory = ?1 ORDER by id DESC;
-- name: UpdateConversationTitle :one
UPDATE conversations SET title = ?2, updated_at = CURRENT_TIMESTAMP WHERE id = ?1 RETURNING *;

-- name: DeleteConversation :execrows
DELETE FROM conversations WHERE id = ?1;
