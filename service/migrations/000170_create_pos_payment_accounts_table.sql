-- +goose Up
CREATE TABLE pos_payment_accounts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  method TEXT NOT NULL,
  account_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_payment_accounts_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_payment_accounts_account_id
    FOREIGN KEY (account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE UNIQUE INDEX idx_pos_payment_accounts_org_method ON pos_payment_accounts(organization_id, method);
CREATE INDEX idx_pos_payment_accounts_organization_id ON pos_payment_accounts(organization_id);

-- +goose Down
DROP TABLE pos_payment_accounts;
