-- +goose Up
SELECT 'up SQL query';
CREATE TABLE journal_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY,
  movement_id BIGINT NOT NULL,
  date DATE NOT NULL,
  account_id BIGINT NOT NULL,
  contact_id BIGINT,
  name TEXT,
  debit NUMERIC(18,4) DEFAULT 0,
  credit NUMERIC(18,4) DEFAULT 0,
  currency_code CHAR(3),
  amount_currency NUMERIC(18,4),
  dimension_id BIGINT,
  tax_id BIGINT,
  reconciled BOOLEAN DEFAULT false,
  full_reconcile_id BIGINT,
  due_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (date, id),
  CONSTRAINT journal_lines_debit_credit CHECK (debit >= 0 AND credit >= 0),
  CONSTRAINT journal_lines_not_both CHECK (NOT (debit > 0 AND credit > 0)),
  CONSTRAINT fk_journal_lines_account_id
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_lines_contact_id
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_lines_currency_code
    FOREIGN KEY (currency_code) REFERENCES currencies(code) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_lines_dimension_id
    FOREIGN KEY (dimension_id) REFERENCES dimensions(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_lines_tax_id
    FOREIGN KEY (tax_id) REFERENCES taxes(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_lines_full_reconcile_id
    FOREIGN KEY (full_reconcile_id) REFERENCES account_full_reconciles(id) ON DELETE RESTRICT
) PARTITION BY RANGE (date);

CREATE TABLE journal_lines_default PARTITION OF journal_lines DEFAULT;
CREATE TABLE journal_lines_202001 PARTITION OF journal_lines FOR VALUES FROM ('2020-01-01') TO ('2020-02-01');
CREATE TABLE journal_lines_202002 PARTITION OF journal_lines FOR VALUES FROM ('2020-02-01') TO ('2020-03-01');
CREATE TABLE journal_lines_202003 PARTITION OF journal_lines FOR VALUES FROM ('2020-03-01') TO ('2020-04-01');
CREATE TABLE journal_lines_202004 PARTITION OF journal_lines FOR VALUES FROM ('2020-04-01') TO ('2020-05-01');
CREATE TABLE journal_lines_202005 PARTITION OF journal_lines FOR VALUES FROM ('2020-05-01') TO ('2020-06-01');
CREATE TABLE journal_lines_202006 PARTITION OF journal_lines FOR VALUES FROM ('2020-06-01') TO ('2020-07-01');
CREATE TABLE journal_lines_202007 PARTITION OF journal_lines FOR VALUES FROM ('2020-07-01') TO ('2020-08-01');
CREATE TABLE journal_lines_202008 PARTITION OF journal_lines FOR VALUES FROM ('2020-08-01') TO ('2020-09-01');
CREATE TABLE journal_lines_202009 PARTITION OF journal_lines FOR VALUES FROM ('2020-09-01') TO ('2020-10-01');
CREATE TABLE journal_lines_202010 PARTITION OF journal_lines FOR VALUES FROM ('2020-10-01') TO ('2020-11-01');
CREATE TABLE journal_lines_202011 PARTITION OF journal_lines FOR VALUES FROM ('2020-11-01') TO ('2020-12-01');
CREATE TABLE journal_lines_202012 PARTITION OF journal_lines FOR VALUES FROM ('2020-12-01') TO ('2021-01-01');
CREATE TABLE journal_lines_202101 PARTITION OF journal_lines FOR VALUES FROM ('2021-01-01') TO ('2021-02-01');
CREATE TABLE journal_lines_202102 PARTITION OF journal_lines FOR VALUES FROM ('2021-02-01') TO ('2021-03-01');
CREATE TABLE journal_lines_202103 PARTITION OF journal_lines FOR VALUES FROM ('2021-03-01') TO ('2021-04-01');
CREATE TABLE journal_lines_202104 PARTITION OF journal_lines FOR VALUES FROM ('2021-04-01') TO ('2021-05-01');
CREATE TABLE journal_lines_202105 PARTITION OF journal_lines FOR VALUES FROM ('2021-05-01') TO ('2021-06-01');
CREATE TABLE journal_lines_202106 PARTITION OF journal_lines FOR VALUES FROM ('2021-06-01') TO ('2021-07-01');
CREATE TABLE journal_lines_202107 PARTITION OF journal_lines FOR VALUES FROM ('2021-07-01') TO ('2021-08-01');
CREATE TABLE journal_lines_202108 PARTITION OF journal_lines FOR VALUES FROM ('2021-08-01') TO ('2021-09-01');
CREATE TABLE journal_lines_202109 PARTITION OF journal_lines FOR VALUES FROM ('2021-09-01') TO ('2021-10-01');
CREATE TABLE journal_lines_202110 PARTITION OF journal_lines FOR VALUES FROM ('2021-10-01') TO ('2021-11-01');
CREATE TABLE journal_lines_202111 PARTITION OF journal_lines FOR VALUES FROM ('2021-11-01') TO ('2021-12-01');
CREATE TABLE journal_lines_202112 PARTITION OF journal_lines FOR VALUES FROM ('2021-12-01') TO ('2022-01-01');
CREATE TABLE journal_lines_202201 PARTITION OF journal_lines FOR VALUES FROM ('2022-01-01') TO ('2022-02-01');
CREATE TABLE journal_lines_202202 PARTITION OF journal_lines FOR VALUES FROM ('2022-02-01') TO ('2022-03-01');
CREATE TABLE journal_lines_202203 PARTITION OF journal_lines FOR VALUES FROM ('2022-03-01') TO ('2022-04-01');
CREATE TABLE journal_lines_202204 PARTITION OF journal_lines FOR VALUES FROM ('2022-04-01') TO ('2022-05-01');
CREATE TABLE journal_lines_202205 PARTITION OF journal_lines FOR VALUES FROM ('2022-05-01') TO ('2022-06-01');
CREATE TABLE journal_lines_202206 PARTITION OF journal_lines FOR VALUES FROM ('2022-06-01') TO ('2022-07-01');
CREATE TABLE journal_lines_202207 PARTITION OF journal_lines FOR VALUES FROM ('2022-07-01') TO ('2022-08-01');
CREATE TABLE journal_lines_202208 PARTITION OF journal_lines FOR VALUES FROM ('2022-08-01') TO ('2022-09-01');
CREATE TABLE journal_lines_202209 PARTITION OF journal_lines FOR VALUES FROM ('2022-09-01') TO ('2022-10-01');
CREATE TABLE journal_lines_202210 PARTITION OF journal_lines FOR VALUES FROM ('2022-10-01') TO ('2022-11-01');
CREATE TABLE journal_lines_202211 PARTITION OF journal_lines FOR VALUES FROM ('2022-11-01') TO ('2022-12-01');
CREATE TABLE journal_lines_202212 PARTITION OF journal_lines FOR VALUES FROM ('2022-12-01') TO ('2023-01-01');
CREATE TABLE journal_lines_202301 PARTITION OF journal_lines FOR VALUES FROM ('2023-01-01') TO ('2023-02-01');
CREATE TABLE journal_lines_202302 PARTITION OF journal_lines FOR VALUES FROM ('2023-02-01') TO ('2023-03-01');
CREATE TABLE journal_lines_202303 PARTITION OF journal_lines FOR VALUES FROM ('2023-03-01') TO ('2023-04-01');
CREATE TABLE journal_lines_202304 PARTITION OF journal_lines FOR VALUES FROM ('2023-04-01') TO ('2023-05-01');
CREATE TABLE journal_lines_202305 PARTITION OF journal_lines FOR VALUES FROM ('2023-05-01') TO ('2023-06-01');
CREATE TABLE journal_lines_202306 PARTITION OF journal_lines FOR VALUES FROM ('2023-06-01') TO ('2023-07-01');
CREATE TABLE journal_lines_202307 PARTITION OF journal_lines FOR VALUES FROM ('2023-07-01') TO ('2023-08-01');
CREATE TABLE journal_lines_202308 PARTITION OF journal_lines FOR VALUES FROM ('2023-08-01') TO ('2023-09-01');
CREATE TABLE journal_lines_202309 PARTITION OF journal_lines FOR VALUES FROM ('2023-09-01') TO ('2023-10-01');
CREATE TABLE journal_lines_202310 PARTITION OF journal_lines FOR VALUES FROM ('2023-10-01') TO ('2023-11-01');
CREATE TABLE journal_lines_202311 PARTITION OF journal_lines FOR VALUES FROM ('2023-11-01') TO ('2023-12-01');
CREATE TABLE journal_lines_202312 PARTITION OF journal_lines FOR VALUES FROM ('2023-12-01') TO ('2024-01-01');
CREATE TABLE journal_lines_202401 PARTITION OF journal_lines FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE journal_lines_202402 PARTITION OF journal_lines FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE journal_lines_202403 PARTITION OF journal_lines FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
CREATE TABLE journal_lines_202404 PARTITION OF journal_lines FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
CREATE TABLE journal_lines_202405 PARTITION OF journal_lines FOR VALUES FROM ('2024-05-01') TO ('2024-06-01');
CREATE TABLE journal_lines_202406 PARTITION OF journal_lines FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');
CREATE TABLE journal_lines_202407 PARTITION OF journal_lines FOR VALUES FROM ('2024-07-01') TO ('2024-08-01');
CREATE TABLE journal_lines_202408 PARTITION OF journal_lines FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
CREATE TABLE journal_lines_202409 PARTITION OF journal_lines FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');
CREATE TABLE journal_lines_202410 PARTITION OF journal_lines FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE journal_lines_202411 PARTITION OF journal_lines FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE journal_lines_202412 PARTITION OF journal_lines FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');
CREATE TABLE journal_lines_202501 PARTITION OF journal_lines FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE journal_lines_202502 PARTITION OF journal_lines FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE journal_lines_202503 PARTITION OF journal_lines FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE journal_lines_202504 PARTITION OF journal_lines FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE journal_lines_202505 PARTITION OF journal_lines FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE journal_lines_202506 PARTITION OF journal_lines FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE journal_lines_202507 PARTITION OF journal_lines FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE journal_lines_202508 PARTITION OF journal_lines FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE journal_lines_202509 PARTITION OF journal_lines FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE journal_lines_202510 PARTITION OF journal_lines FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE journal_lines_202511 PARTITION OF journal_lines FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE journal_lines_202512 PARTITION OF journal_lines FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE journal_lines_202601 PARTITION OF journal_lines FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE journal_lines_202602 PARTITION OF journal_lines FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE journal_lines_202603 PARTITION OF journal_lines FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE journal_lines_202604 PARTITION OF journal_lines FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE journal_lines_202605 PARTITION OF journal_lines FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE journal_lines_202606 PARTITION OF journal_lines FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE journal_lines_202607 PARTITION OF journal_lines FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE journal_lines_202608 PARTITION OF journal_lines FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE journal_lines_202609 PARTITION OF journal_lines FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE journal_lines_202610 PARTITION OF journal_lines FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE journal_lines_202611 PARTITION OF journal_lines FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE journal_lines_202612 PARTITION OF journal_lines FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE journal_lines_202701 PARTITION OF journal_lines FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE journal_lines_202702 PARTITION OF journal_lines FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE journal_lines_202703 PARTITION OF journal_lines FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE journal_lines_202704 PARTITION OF journal_lines FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE journal_lines_202705 PARTITION OF journal_lines FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE journal_lines_202706 PARTITION OF journal_lines FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE journal_lines_202707 PARTITION OF journal_lines FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE journal_lines_202708 PARTITION OF journal_lines FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE journal_lines_202709 PARTITION OF journal_lines FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE journal_lines_202710 PARTITION OF journal_lines FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE journal_lines_202711 PARTITION OF journal_lines FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE journal_lines_202712 PARTITION OF journal_lines FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');
CREATE TABLE journal_lines_202801 PARTITION OF journal_lines FOR VALUES FROM ('2028-01-01') TO ('2028-02-01');
CREATE TABLE journal_lines_202802 PARTITION OF journal_lines FOR VALUES FROM ('2028-02-01') TO ('2028-03-01');
CREATE TABLE journal_lines_202803 PARTITION OF journal_lines FOR VALUES FROM ('2028-03-01') TO ('2028-04-01');
CREATE TABLE journal_lines_202804 PARTITION OF journal_lines FOR VALUES FROM ('2028-04-01') TO ('2028-05-01');
CREATE TABLE journal_lines_202805 PARTITION OF journal_lines FOR VALUES FROM ('2028-05-01') TO ('2028-06-01');
CREATE TABLE journal_lines_202806 PARTITION OF journal_lines FOR VALUES FROM ('2028-06-01') TO ('2028-07-01');
CREATE TABLE journal_lines_202807 PARTITION OF journal_lines FOR VALUES FROM ('2028-07-01') TO ('2028-08-01');
CREATE TABLE journal_lines_202808 PARTITION OF journal_lines FOR VALUES FROM ('2028-08-01') TO ('2028-09-01');
CREATE TABLE journal_lines_202809 PARTITION OF journal_lines FOR VALUES FROM ('2028-09-01') TO ('2028-10-01');
CREATE TABLE journal_lines_202810 PARTITION OF journal_lines FOR VALUES FROM ('2028-10-01') TO ('2028-11-01');
CREATE TABLE journal_lines_202811 PARTITION OF journal_lines FOR VALUES FROM ('2028-11-01') TO ('2028-12-01');
CREATE TABLE journal_lines_202812 PARTITION OF journal_lines FOR VALUES FROM ('2028-12-01') TO ('2029-01-01');
CREATE TABLE journal_lines_202901 PARTITION OF journal_lines FOR VALUES FROM ('2029-01-01') TO ('2029-02-01');
CREATE TABLE journal_lines_202902 PARTITION OF journal_lines FOR VALUES FROM ('2029-02-01') TO ('2029-03-01');
CREATE TABLE journal_lines_202903 PARTITION OF journal_lines FOR VALUES FROM ('2029-03-01') TO ('2029-04-01');
CREATE TABLE journal_lines_202904 PARTITION OF journal_lines FOR VALUES FROM ('2029-04-01') TO ('2029-05-01');
CREATE TABLE journal_lines_202905 PARTITION OF journal_lines FOR VALUES FROM ('2029-05-01') TO ('2029-06-01');
CREATE TABLE journal_lines_202906 PARTITION OF journal_lines FOR VALUES FROM ('2029-06-01') TO ('2029-07-01');
CREATE TABLE journal_lines_202907 PARTITION OF journal_lines FOR VALUES FROM ('2029-07-01') TO ('2029-08-01');
CREATE TABLE journal_lines_202908 PARTITION OF journal_lines FOR VALUES FROM ('2029-08-01') TO ('2029-09-01');
CREATE TABLE journal_lines_202909 PARTITION OF journal_lines FOR VALUES FROM ('2029-09-01') TO ('2029-10-01');
CREATE TABLE journal_lines_202910 PARTITION OF journal_lines FOR VALUES FROM ('2029-10-01') TO ('2029-11-01');
CREATE TABLE journal_lines_202911 PARTITION OF journal_lines FOR VALUES FROM ('2029-11-01') TO ('2029-12-01');
CREATE TABLE journal_lines_202912 PARTITION OF journal_lines FOR VALUES FROM ('2029-12-01') TO ('2030-01-01');
CREATE TABLE journal_lines_203001 PARTITION OF journal_lines FOR VALUES FROM ('2030-01-01') TO ('2030-02-01');
CREATE TABLE journal_lines_203002 PARTITION OF journal_lines FOR VALUES FROM ('2030-02-01') TO ('2030-03-01');
CREATE TABLE journal_lines_203003 PARTITION OF journal_lines FOR VALUES FROM ('2030-03-01') TO ('2030-04-01');
CREATE TABLE journal_lines_203004 PARTITION OF journal_lines FOR VALUES FROM ('2030-04-01') TO ('2030-05-01');
CREATE TABLE journal_lines_203005 PARTITION OF journal_lines FOR VALUES FROM ('2030-05-01') TO ('2030-06-01');
CREATE TABLE journal_lines_203006 PARTITION OF journal_lines FOR VALUES FROM ('2030-06-01') TO ('2030-07-01');
CREATE TABLE journal_lines_203007 PARTITION OF journal_lines FOR VALUES FROM ('2030-07-01') TO ('2030-08-01');
CREATE TABLE journal_lines_203008 PARTITION OF journal_lines FOR VALUES FROM ('2030-08-01') TO ('2030-09-01');
CREATE TABLE journal_lines_203009 PARTITION OF journal_lines FOR VALUES FROM ('2030-09-01') TO ('2030-10-01');
CREATE TABLE journal_lines_203010 PARTITION OF journal_lines FOR VALUES FROM ('2030-10-01') TO ('2030-11-01');
CREATE TABLE journal_lines_203011 PARTITION OF journal_lines FOR VALUES FROM ('2030-11-01') TO ('2030-12-01');
CREATE TABLE journal_lines_203012 PARTITION OF journal_lines FOR VALUES FROM ('2030-12-01') TO ('2031-01-01');
CREATE INDEX idx_aml_move ON journal_lines(movement_id);
CREATE INDEX idx_aml_account ON journal_lines(account_id);
CREATE INDEX idx_aml_contact_open ON journal_lines(contact_id) WHERE reconciled = false;
CREATE INDEX idx_aml_id ON journal_lines(id);
CREATE INDEX idx_aml_reconcile ON journal_lines(full_reconcile_id) WHERE full_reconcile_id IS NOT NULL;
CREATE INDEX idx_aml_reconciled ON journal_lines(account_id, reconciled) WHERE reconciled = false;
CREATE INDEX idx_aml_currency ON journal_lines(currency_code) WHERE currency_code IS NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_journal_lines_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS sw_partition_moving(tbl name NOT NULL, id bigint NOT NULL, PRIMARY KEY (tbl, id)) ON COMMIT DELETE ROWS;
  IF EXISTS (SELECT 1 FROM sw_partition_moving WHERE tbl = TG_TABLE_NAME AND id = OLD.id) THEN
    RETURN OLD;
  END IF;
  DELETE FROM account_partial_reconciles WHERE credit_line_id = OLD.id OR debit_line_id = OLD.id;
  IF EXISTS (SELECT 1 FROM bank_statement_lines WHERE journal_line_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move line %: referenced by bank_statement_lines', OLD.id;
  END IF;
  RETURN OLD;
END $$;
-- +goose StatementEnd

CREATE TRIGGER enforce_aml_move BEFORE INSERT OR UPDATE OF movement_id, date ON journal_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_aml_move();
CREATE TRIGGER enforce_aml_delete BEFORE DELETE ON journal_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_journal_lines_delete();
CREATE TRIGGER register_aml_move BEFORE UPDATE OF date ON journal_lines
  FOR EACH ROW EXECUTE FUNCTION register_partition_move();
CREATE TRIGGER clear_aml_move AFTER INSERT ON journal_lines
  FOR EACH ROW EXECUTE FUNCTION clear_partition_move();

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_aml_move ON journal_lines;
DROP TRIGGER IF EXISTS enforce_aml_delete ON journal_lines;
DROP TRIGGER IF EXISTS register_aml_move ON journal_lines;
DROP TRIGGER IF EXISTS clear_aml_move ON journal_lines;
DROP FUNCTION IF EXISTS enforce_journal_lines_delete();
DROP INDEX IF EXISTS idx_aml_move;
DROP INDEX IF EXISTS idx_aml_account;
DROP INDEX IF EXISTS idx_aml_contact_open;
DROP INDEX IF EXISTS idx_aml_id;
DROP TABLE journal_lines;
