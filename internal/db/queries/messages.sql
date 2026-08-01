-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    role,
    content,
    error
) VALUES (
    ?1, ?2, ?3, ?4
) RETURNING id;

-- name: GetMessagesByConversation :many
SELECT id, conversation_id, role, content, error, created_at
FROM messages
WHERE conversation_id = ?1
ORDER BY id ASC;

-- name: CreateToolCall :one
INSERT INTO tool_calls (
    message_id,
    call_id,
    name,
    arguments,
    result,
    is_error,
    status
) VALUES (
    ?1, ?2, ?3, ?4, ?5, ?6, ?7
) RETURNING id;

-- name: UpdateToolResult :exec
UPDATE tool_calls
SET result = ?1,
    is_error = ?2,
    status = ?3
WHERE call_id = ?4;

-- name: GetToolCallsByConversation :many
SELECT tc.id, tc.message_id, tc.call_id, tc.name, tc.arguments,
       tc.result, tc.is_error, tc.status, tc.created_at
FROM tool_calls tc
JOIN messages m ON m.id = tc.message_id
WHERE m.conversation_id = ?1
ORDER BY tc.id ASC;

-- name: GetToolCallsByMessage :many
SELECT id, message_id, call_id, name, arguments,
       result, is_error, status, created_at
FROM tool_calls
WHERE message_id = ?1
ORDER BY id ASC;
