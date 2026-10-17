-- name: AddToken :one
INSERT INTO refresh_tokens (user_id,created_at,updated_at, token, expires_at,revoked_at) 
VALUES ($1, NOW(), NOW(), $2, $3, $4) RETURNING token, created_at, updated_at, expires_at, revoked_at;

-- name: GetToken :one
SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens 
SET updated_at = NOW(), revoked_at = NOW()
WHERE $1 = token;