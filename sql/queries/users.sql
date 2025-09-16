-- name: CreateUser :one
INSERT INTO users(id, email, name, created_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW()
)
RETURNING *;