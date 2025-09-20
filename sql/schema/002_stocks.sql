-- +goose Up
CREATE TABLE stocks(
    symbol TEXT PRIMARY KEY,
    company_name TEXT NOT NULL,
    current_price DOUBLE PRECISION NOT NULL,
    previous_close DOUBLE PRECISION,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE stocks;