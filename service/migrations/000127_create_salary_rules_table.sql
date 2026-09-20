-- +goose Up
SELECT 'up SQL query';
CREATE TABLE salary_rules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  code TEXT,
  name TEXT,
  category TEXT,
  compute_type TEXT,
  amount NUMERIC(18,4),
  formula TEXT,
  account_debit_id BIGINT,
  account_credit_id BIGINT,
  organization_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_salary_rules_account_debit_id
    FOREIGN KEY (account_debit_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_salary_rules_account_credit_id
    FOREIGN KEY (account_credit_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_salary_rules_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE salary_rules;
