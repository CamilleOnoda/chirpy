-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserFromRefreshToken :one
SELECT users.* 
FROM refresh_tokens
JOIN users ON refresh_tokens.user_id = users.id
WHERE refresh_tokens.token = $1 
AND revoked_at IS NULL 
AND expires_at > NOW();
