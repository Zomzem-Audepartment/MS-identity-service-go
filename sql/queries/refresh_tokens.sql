-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token, expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;

-- name: RevokeUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
