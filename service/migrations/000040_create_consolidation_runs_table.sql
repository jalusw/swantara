-- +goose Up
SELECT 'up SQL query';
CREATE TABLE consolidation_runs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  group_organization_id BIGINT,
  period_id BIGINT,
  reporting_currency CHAR(3),
  state TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_consolidation_runs_group_organization_id
    FOREIGN KEY (group_organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_consolidation_runs_period_id
    FOREIGN KEY (period_id)
    REFERENCES tax_periods(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_consolidation_runs_reporting_currency
    FOREIGN KEY (reporting_currency)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE consolidation_runs;
