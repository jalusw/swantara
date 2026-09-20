-- +goose Up
SELECT 'up SQL query';
CREATE TABLE fx_rates (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  currency_code CHAR(3) NOT NULL,
  organization_id BIGINT,
  rate NUMERIC(18,8) NOT NULL,
  rate_type TEXT DEFAULT 'spot',
  valid_from DATE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_fx_rates_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_fx_rates_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  UNIQUE (currency_code, organization_id, rate_type, valid_from)
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE fx_rates;
