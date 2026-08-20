-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    role,
    parts,
    error
) VALUES (
    ?1, ?2, ?3, ?4
) RETURNING id;

-- name: GetMessagesByConversation :many
SELECT *
FROM messages
WHERE conversation_id = ?1
ORDER BY id ASC;
