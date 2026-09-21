-- name: CreateChirps :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (gen_random_uuid(),
        now(),
        now(),
        $1,
        $2)
RETURNING *;

-- name: DeleteAllChirps :exec
DELETE
FROM chirps;

-- name: GetAllChirps :many
SELECT *
from chirps
order by created_at;

-- name: GetChirp :one
SELECT *
FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE
FROM chirps
WHERE id = $1;
