- name: AddOrUpdateProvider :exec
INSERT INTO provider (id , name, provider_id, model_id) VALUES (1 , ?1, ?2, ?3) ON CONFLICT (id) DO UPDATE SET name = ?1, provider_id = ?2, model_id = ?3;

-- name: GetProvider :one
SELECT * FROM provider WHERE id = 1;
