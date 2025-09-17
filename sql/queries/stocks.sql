-- name: CreateNewStock :one
INSERT INTO stocks(symbol, company_name, current_price, previous_close, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    NOW()
)
RETURNING *;

-- name: GetStockBySymbol :one
SELECT * FROM stocks
WHERE symbol = $1;

-- name: UpdateStockPrice :one
UPDATE stocks
SET 
    current_price = $1,
    previous_close = $2,
    updated_at = NOW()
WHERE symbol = $3
RETURNING *;

-- name: SearchStockByName :many
SELECT *
FROM stocks
WHERE company_name ILIKE '%' || $1 || '%';

-- name: GetAllStocks :many
SELECT * FROM stocks;