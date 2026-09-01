-- name: AddOrUpdateSettings :exec
INSERT 
	INTO user_settings (id, theme) 
	VALUES (1, ?1) 
		ON CONFLICT (id) 
		DO UPDATE 
		SET theme = ?1;

-- name: GetSettings :one
SELECT * FROM user_settings WHERE id = 1;
