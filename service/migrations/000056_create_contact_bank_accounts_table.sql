-- +goose Up
SELECT 'up SQL query';
CREATE TABLE contact_bank_accounts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  contact_id BIGINT NOT NULL,
  account_holder TEXT,
  bank_name TEXT,
  iban TEXT,
  swift_bic TEXT,
  account_number TEXT,
  routing_number TEXT,
  currency_code CHAR(3),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_contact_bank_accounts_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_contact_bank_accounts_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE contact_bank_accounts;
