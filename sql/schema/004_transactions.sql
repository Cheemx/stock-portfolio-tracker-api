-- +goose Up
CREATE TABLE transactions(
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    stock_symbol TEXT REFERENCES stocks(symbol) ON DELETE CASCADE NOT NULL,
    type TEXT CHECK (type IN ('BUY', 'SELL')) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE transactions;