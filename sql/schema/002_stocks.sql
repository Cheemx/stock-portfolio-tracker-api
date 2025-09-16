-- +goose Up
CREATE TABLE stocks(
    symbol TEXT PRIMARY KEY,
    company_name TEXT NOT NULL,
    current_price DECIMAL(10, 2) NOT NULL,
    previous_close DECIMAL(10, 2),
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE stocks;