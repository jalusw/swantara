-- +goose Up
SELECT 'up SQL query';
CREATE TABLE commission_entries (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  salesperson_id BIGINT,
  plan_id BIGINT,
  source_type TEXT,
  source_id BIGINT,
  base_amount NUMERIC(18,4),
  commission_amount NUMERIC(18,4),
  state TEXT CHECK (state IN ('draft','confirmed','paid','cancelled')),
  period_id BIGINT,
  payslip_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_commission_entries_salesperson_id
    FOREIGN KEY (salesperson_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_commission_entries_plan_id
    FOREIGN KEY (plan_id)
    REFERENCES commission_plans(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_commission_entries_period_id
    FOREIGN KEY (period_id)
    REFERENCES tax_periods(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_commission_entries_payslip_id
    FOREIGN KEY (payslip_id)
    REFERENCES payslips(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE commission_entries;
