-- +goose Up
SELECT 'up SQL query';
CREATE TABLE expense_reports (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  employee_id BIGINT NOT NULL,
  state TEXT CHECK (state IN ('draft','submitted','approved','refused','posted','reimbursed')),
  payment_mode TEXT CHECK (payment_mode IN ('own_account','organization_account')),
  total_amount NUMERIC(18,4),
  movement_id BIGINT,
  reimbursement_movement_id BIGINT,
  submitted_at TIMESTAMP,
  approved_by BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_expense_reports_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_reports_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_er_move BEFORE INSERT OR UPDATE OF movement_id ON expense_reports
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');
CREATE TRIGGER enforce_er_reimbursement BEFORE INSERT OR UPDATE OF reimbursement_movement_id ON expense_reports
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'reimbursement_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_er_reimbursement ON expense_reports;
DROP TRIGGER IF EXISTS enforce_er_move ON expense_reports;
DROP TABLE expense_reports;
