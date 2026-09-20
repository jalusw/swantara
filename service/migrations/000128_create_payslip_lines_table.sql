-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payslip_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  payslip_id BIGINT,
  rule_id BIGINT,
  code TEXT,
  name TEXT,
  category TEXT,
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payslip_lines_payslip_id
    FOREIGN KEY (payslip_id)
    REFERENCES payslips(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payslip_lines_rule_id
    FOREIGN KEY (rule_id)
    REFERENCES salary_rules(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE payslip_lines;
