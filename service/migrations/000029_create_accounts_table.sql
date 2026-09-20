-- +goose Up
SELECT 'up SQL query';
CREATE TABLE accounts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  type TEXT CHECK (type IN ('asset','liability','equity','income','expense',
                              'receivable','payable','bank','cash','cogs','tax',
                              'current_asset','fixed_asset','depreciation')),
  reconcilable BOOLEAN DEFAULT false,
  currency_code CHAR(3),
  parent_id BIGINT,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_accounts_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_accounts_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_accounts_parent_id
    FOREIGN KEY (parent_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  UNIQUE (organization_id, code)
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE accounts;
