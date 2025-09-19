-- name: CreateUser :one
INSERT INTO users(id, email, name, created_at, hashed_password)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    $3
)
RETURNING *;

-- name: DeleteAllUsers :exec
TRUNCATE TABLE users CASCADE;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;