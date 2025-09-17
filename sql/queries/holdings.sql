-- name: GetAllHoldingsForUser :many
SELECT 
    holdings.stock_symbol AS stock_symbol,
    stocks.company_name AS company_name,
    holdings.quantity AS quantity,
    holdings.average_price AS average_price,
    stocks.current_price AS current_price,
    (holdings.quantity * stocks.current_price) AS current_value
FROM holdings
JOIN stocks
ON holdings.stock_symbol = stocks.symbol
WHERE holdings.user_id = $1;

-- name: DeleteHoldingsOnSellOut :execrows
DELETE FROM holdings 
WHERE holdings.user_id = $1 AND holdings.stock_symbol = $2;

-- name: CreateHoldingForUser :one
INSERT INTO holdings(id, user_id, stock_symbol, quantity, average_price, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    NOW(),
    NOW()
)
RETURNING *;

-- name: UpdateHoldingOnTransaction :one
UPDATE holdings
SET 
    quantity = $1,
    average_price = $2,
    updated_at = NOW()
WHERE user_id = $3 AND stock_symbol = $4
RETURNING *;

-- name: GetHoldingByStockSymbol :one
SELECT * FROM holdings
WHERE user_id = $1 AND stock_symbol = $2;