-- +goose Up
SELECT 'up SQL query';
CREATE TABLE taxes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT NOT NULL,
  amount NUMERIC(8,4),
  type TEXT CHECK (type IN ('percent','fixed','group')),
  scope TEXT CHECK (scope IN ('sale','purchase','none')),
  price_include BOOLEAN DEFAULT false,
  tax_account_id BIGINT,
  refund_tax_account_id BIGINT,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_taxes_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_taxes_tax_account_id
    FOREIGN KEY (tax_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_taxes_refund_tax_account_id
    FOREIGN KEY (refund_tax_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE taxes;
