-- +goose Up
SELECT 'up SQL query';
CREATE TABLE tax_periods (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  tax_year_id BIGINT,
  name TEXT,
  date_start DATE,
  date_end DATE,
  state TEXT CHECK (state IN ('open','closed','locked')),
  period_type TEXT DEFAULT 'standard' CHECK (period_type IN ('standard','adjustment')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_tax_periods_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_tax_periods_tax_year_id
    FOREIGN KEY (tax_year_id)
    REFERENCES tax_years(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_tax_periods_type ON tax_periods(period_type);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE tax_periods;
