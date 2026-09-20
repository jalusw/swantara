-- +goose Up
SELECT 'up SQL query';
CREATE TABLE journals (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  code TEXT,
  type TEXT CHECK (type IN ('sale','purchase','bank','cash','general')),
  default_account_id BIGINT,
  currency_code CHAR(3),
  bank_account_id BIGINT,
  sequence_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_journals_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_journals_default_account_id
    FOREIGN KEY (default_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_journals_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_journals_bank_account_id
    FOREIGN KEY (bank_account_id)
    REFERENCES contact_bank_accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_journals_sequence_id
    FOREIGN KEY (sequence_id)
    REFERENCES doc_sequences(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE journals;
