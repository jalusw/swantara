-- +goose Up
SELECT 'up SQL query';
CREATE TABLE invoice_taxes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  invoice_id BIGINT,
  tax_id BIGINT,
  base NUMERIC(18,4),
  amount NUMERIC(18,4),
  account_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoice_taxes_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_taxes_tax_id
    FOREIGN KEY (tax_id)
    REFERENCES taxes(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE invoice_taxes;
