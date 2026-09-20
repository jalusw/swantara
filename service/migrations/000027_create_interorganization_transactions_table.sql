-- +goose Up
SELECT 'up SQL query';
CREATE TABLE interorganization_transactions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_organization_id BIGINT,
  source_type TEXT,
  source_id BIGINT,
  mirror_organization_id BIGINT,
  mirror_type TEXT,
  mirror_id BIGINT,
  amount NUMERIC(18,4),
  state TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE interorganization_transactions;
