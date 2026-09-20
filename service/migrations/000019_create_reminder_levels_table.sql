-- +goose Up
SELECT 'up SQL query';
CREATE TABLE reminder_levels (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  days_overdue INT,
  sequence INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE reminder_levels;
