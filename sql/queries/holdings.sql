-- name: GetAllHoldingsForUser :many
SELECT 
    holdings.stock_symbol AS stock_symbol,
    stocks.company_name AS company_name,
    holdings.quantity AS quantity,
    holdings.average_price AS average_price,
    stocks.current_price AS current_price,
    holdings.total_invested AS total_invested
FROM holdings
JOIN stocks
ON holdings.stock_symbol = stocks.symbol
WHERE holdings.user_id = $1;

-- name: DeleteHoldingsOnSellOut :execrows
DELETE FROM holdings 
WHERE holdings.user_id = $1 AND holdings.stock_symbol = $2;

-- name: CreateNewHoldingOrUpdateExistingForUser :one
INSERT INTO holdings(id, user_id, stock_symbol, quantity, average_price, created_at, updated_at, total_invested)
VALUES (
    gen_random_uuid(),
    $1,  
    $2,  
    $3,  
    $4,  
    NOW(),
    NOW(),
    $5
)
ON CONFLICT (user_id, stock_symbol) DO UPDATE
SET 
    quantity      = EXCLUDED.quantity,
    average_price = EXCLUDED.average_price,
    updated_at    = NOW(),
    total_invested = EXCLUDED.total_invested
RETURNING *;

-- name: GetHoldingByStockSymbol :one
SELECT * FROM holdings
WHERE user_id = $1 AND stock_symbol = $2;

-- name: GetStockSymbolsOfHoldings :many
SELECT DISTINCT stock_symbol
FROM holdings;