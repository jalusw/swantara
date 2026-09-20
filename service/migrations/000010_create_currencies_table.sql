-- +goose Up
SELECT 'up SQL query';
CREATE TABLE currencies (
  id BIGSERIAL PRIMARY KEY,
  code CHAR(3) NOT NULL UNIQUE,
  name TEXT NOT NULL,
  symbol TEXT,
  decimal_places SMALLINT DEFAULT 2,
  rounding NUMERIC(12,6) DEFAULT 0.01,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE currencies;