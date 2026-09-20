-- +goose Up
SELECT 'up SQL query';
CREATE TABLE audit_logs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  table_name TEXT,
  record_id BIGINT,
  action TEXT CHECK (action IN ('insert','update','delete')),
  changed_by BIGINT,
  changed_at TIMESTAMP,
  diff JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE audit_logs;
