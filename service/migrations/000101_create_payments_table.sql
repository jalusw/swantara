-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  contact_id BIGINT,
  type TEXT CHECK (type IN ('inbound','outbound')),
  journal_id BIGINT,
  payment_method TEXT,
  amount NUMERIC(18,4) NOT NULL,
  currency_code CHAR(3),
  date DATE NOT NULL,
  reference TEXT,
  movement_id BIGINT,
  contact_bank_account_id BIGINT,
  state TEXT CHECK (state IN ('draft','posted','reconciled','cancelled')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payments_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payments_journal_id
    FOREIGN KEY (journal_id)
    REFERENCES journals(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payments_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payments_contact_bank_account_id
    FOREIGN KEY (contact_bank_account_id)
    REFERENCES contact_bank_accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_pay_move BEFORE INSERT OR UPDATE OF movement_id ON payments
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_pay_move ON payments;
DROP TABLE payments;
