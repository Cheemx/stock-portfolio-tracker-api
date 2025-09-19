-- name: GetAllTransactionsForUser :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC 
LIMIT 10;

-- name: CreateATransaction :one
INSERT INTO transactions(id, user_id, stock_symbol, type, quantity, price, total_amount, created_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    NOW()
)
RETURNING *;

-- name: GetAllTransactionsForUserBySymbol :many
SELECT * FROM transactions
WHERE user_id = $1 AND stock_symbol = $2;