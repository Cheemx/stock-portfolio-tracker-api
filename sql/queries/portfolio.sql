-- name: GetPortfolioForUser :one
SELECT 
        SUM(holdings.total_invested)::DOUBLE PRECISION AS total_invested,
        SUM(holdings.quantity * stocks.current_price)::DOUBLE PRECISION AS current_value,
        COUNT(holdings.user_id) AS holdings_count
FROM holdings 
JOIN users
ON holdings.user_id = users.id
JOIN stocks 
ON holdings.stock_symbol = stocks.symbol
WHERE users.id = $1
GROUP BY holdings.user_id;