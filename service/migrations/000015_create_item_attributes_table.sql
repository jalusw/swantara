-- +goose Up
SELECT 'up SQL query';
CREATE TABLE item_attributes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE item_attributes;
