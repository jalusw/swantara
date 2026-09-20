-- +goose Up
SELECT 'up SQL query';
CREATE TABLE integration_events (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  topic TEXT,
  payload JSONB,
  status TEXT CHECK (status IN ('pending','sent','failed')),
  retries INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE integration_events;
