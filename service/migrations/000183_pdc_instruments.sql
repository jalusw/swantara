-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pdc_instruments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  contact_id BIGINT NOT NULL,
  direction TEXT CHECK (direction IN ('inbound','outbound')),
  number TEXT,
  bank_name TEXT,
  amount NUMERIC(18,4),
  currency_code CHAR(3),
  due_date DATE,
  state TEXT CHECK (state IN ('held','deposited','cleared','bounced','cancelled')),
  invoice_id BIGINT,
  payment_id BIGINT,
  journal_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pdc_instruments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pdc_instruments_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pdc_instruments_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pdc_instruments_payment_id
    FOREIGN KEY (payment_id)
    REFERENCES payments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pdc_instruments_journal_id
    FOREIGN KEY (journal_id)
    REFERENCES journals(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_pdc_instruments_due ON pdc_instruments(due_date) WHERE state IN ('held','deposited');

-- +goose Down
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_pdc_instruments_due;
DROP TABLE pdc_instruments;
