-- name: AddOrUpdateProvider :exec
INSERT 
	INTO provider (id , provider_name, model_id) 
	VALUES (1 , ?1, ?2) 
		ON CONFLICT (id) 
		DO UPDATE 
		SET provider_name = ?1, model_id = ?2;

-- name: GetProvider :one
SELECT * FROM provider WHERE id = 1;
