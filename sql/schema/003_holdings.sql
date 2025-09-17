-- +goose Up
CREATE TABLE holdings(
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    stock_symbol TEXT REFERENCES stocks(symbol) ON DELETE CASCADE,
    UNIQUE (user_id, stock_symbol),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    average_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE holdings;