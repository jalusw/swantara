-- +goose Up
CREATE TABLE idempotency_keys (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  key TEXT,
  resource TEXT,
  status_code INT,
  response JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT uq_idempotency_keys_organization_key UNIQUE (organization_id, key),
  CONSTRAINT fk_idempotency_keys_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
DROP TABLE idempotency_keys;
