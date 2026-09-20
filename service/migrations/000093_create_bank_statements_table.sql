-- +goose Up
SELECT 'up SQL query';
CREATE TABLE bank_statements (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  journal_id BIGINT,
  name TEXT,
  date DATE,
  balance_start NUMERIC(18,4),
  balance_end NUMERIC(18,4),
  state TEXT,
  currency_code CHAR(3),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_bank_statements_journal_id
    FOREIGN KEY (journal_id)
    REFERENCES journals(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE bank_statements;
