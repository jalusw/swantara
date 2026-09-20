-- +goose Up
SELECT 'up SQL query';
CREATE TABLE bank_statement_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  statement_id BIGINT,
  date DATE,
  amount NUMERIC(18,4),
  contact_id BIGINT,
  ref TEXT,
  narration TEXT,
  reconciled BOOLEAN DEFAULT false,
  payment_id BIGINT,
  journal_line_id BIGINT,
  currency_code CHAR(3),
  fee_amount NUMERIC(18,4) DEFAULT 0,
  interest_amount NUMERIC(18,4) DEFAULT 0,
  fee_movement_id BIGINT,
  interest_movement_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_bank_statement_lines_statement_id
    FOREIGN KEY (statement_id)
    REFERENCES bank_statements(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_bank_statement_lines_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_bank_statement_lines_payment_id
    FOREIGN KEY (payment_id)
    REFERENCES payments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_bsl_journal_line BEFORE INSERT OR UPDATE OF journal_line_id ON bank_statement_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_lines', 'journal_line_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_bsl_journal_line ON bank_statement_lines;
DROP TABLE bank_statement_lines;
