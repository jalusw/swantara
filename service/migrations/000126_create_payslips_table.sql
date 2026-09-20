-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payslips (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  run_id BIGINT,
  employee_id BIGINT,
  contract_id BIGINT,
  gross NUMERIC(18,4),
  net NUMERIC(18,4),
  movement_id BIGINT,
  state TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payslips_run_id
    FOREIGN KEY (run_id)
    REFERENCES payroll_runs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payslips_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payslips_contract_id
    FOREIGN KEY (contract_id)
    REFERENCES employment_contracts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_payslip_move BEFORE INSERT OR UPDATE OF movement_id ON payslips
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_payslip_move ON payslips;
DROP TABLE payslips;
