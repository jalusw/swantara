-- +goose Up
SELECT 'up SQL query';
CREATE TABLE tax_returns (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  period_id BIGINT,
  type TEXT,
  output_tax NUMERIC(18,4),
  input_tax NUMERIC(18,4),
  net_payable NUMERIC(18,4),
  state TEXT CHECK (state IN ('draft','filed','paid')),
  filed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_tax_returns_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_tax_returns_period_id
    FOREIGN KEY (period_id)
    REFERENCES tax_periods(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE tax_returns;
