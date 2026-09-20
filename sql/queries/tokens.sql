-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at ,revoked_at)
VALUES ($1,
        now(),
        now(),
        $2,
        now() + INTERVAL '60 DAYS',
        NULL)
RETURNING *;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: GetUserFromRefreshToken :one
SELECT *
FROM users
WHERE id = (
SELECT user_id
FROM refresh_tokens
WHERE 
        token = $1
        AND expires_at > now()
        AND revoked_at IS NULL
);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now(),
    updated_at = now()
WHERE
token = $1;