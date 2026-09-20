-- +goose Up
SELECT 'up SQL query';
CREATE TABLE consolidation_eliminations (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  consolidation_run_id BIGINT,
  account_id BIGINT,
  counterparty_organization_id BIGINT,
  amount NUMERIC(18,4),
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_consolidation_eliminations_consolidation_run_id
    FOREIGN KEY (consolidation_run_id)
    REFERENCES consolidation_runs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_consolidation_eliminations_account_id
    FOREIGN KEY (account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE consolidation_eliminations;
