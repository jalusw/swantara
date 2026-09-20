-- +goose Up
SELECT 'up SQL query';
CREATE TABLE doc_sequences (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  code TEXT,
  prefix TEXT,
  suffix TEXT,
  next_number BIGINT DEFAULT 1,
  padding INT DEFAULT 5,
  reset_period TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_doc_sequences_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  UNIQUE (organization_id, code)
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE doc_sequences;
