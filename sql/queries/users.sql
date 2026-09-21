-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(),
        now(),
        now(),
        $1,
        $2)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
where email = $1;

-- name: DeleteUser :exec
DELETE
FROM users;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
        hashed_password = $3,
        updated_at = now()
WHERE id = $1
RETURNING *;
