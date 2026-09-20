-- +goose Up
SELECT 'up SQL query';
CREATE TABLE account_partial_reconciles (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  debit_line_id BIGINT,
  credit_line_id BIGINT,
  amount NUMERIC(18,4),
  full_reconcile_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_account_partial_reconciles_full_reconcile_id
    FOREIGN KEY (full_reconcile_id)
    REFERENCES account_full_reconciles(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_apr_credit BEFORE INSERT OR UPDATE OF credit_line_id ON account_partial_reconciles
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_lines', 'credit_line_id');
CREATE TRIGGER enforce_apr_debit BEFORE INSERT OR UPDATE OF debit_line_id ON account_partial_reconciles
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_lines', 'debit_line_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_apr_credit ON account_partial_reconciles;
DROP TRIGGER IF EXISTS enforce_apr_debit ON account_partial_reconciles;
DROP TABLE account_partial_reconciles;
