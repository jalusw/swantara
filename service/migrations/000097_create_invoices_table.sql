-- +goose Up
SELECT 'up SQL query';
CREATE TABLE invoices (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  movement_id BIGINT,
  origin_invoice_id BIGINT,
  type TEXT CHECK (type IN ('customer_invoice','customer_credit_note','supplier_bill','supplier_credit_note')),
  contact_id BIGINT NOT NULL,
  name TEXT,
  reference TEXT,
  invoice_date DATE,
  due_date DATE,
  currency_code CHAR(3),
  journal_id BIGINT,
  payment_term_id BIGINT,
  tax_rule_id BIGINT,
  state TEXT CHECK (state IN ('draft','posted','cancelled')),
  payment_state TEXT CHECK (payment_state IN ('not_paid','in_payment','partial','paid','reversed')),
  amount_untaxed NUMERIC(18,4) DEFAULT 0,
  amount_tax NUMERIC(18,4) DEFAULT 0,
  amount_total NUMERIC(18,4) DEFAULT 0,
  amount_residual NUMERIC(18,4) DEFAULT 0,
  tax_rule TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoices_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_origin_invoice_id
    FOREIGN KEY (origin_invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_journal_id
    FOREIGN KEY (journal_id)
    REFERENCES journals(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_payment_term_id
    FOREIGN KEY (payment_term_id)
    REFERENCES payment_terms(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoices_tax_rule_id
    FOREIGN KEY (tax_rule_id)
    REFERENCES tax_rules(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_invoice_open ON invoices(contact_id, payment_state) WHERE payment_state <> 'paid';

CREATE INDEX idx_invoice_due ON invoices(due_date) WHERE payment_state <> 'paid';

CREATE TRIGGER enforce_inv_move BEFORE INSERT OR UPDATE OF movement_id ON invoices
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_inv_move ON invoices;
DROP INDEX IF EXISTS idx_invoice_open;
DROP INDEX IF EXISTS idx_invoice_due;
DROP TABLE invoices;
