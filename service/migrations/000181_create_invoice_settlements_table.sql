-- +goose Up
SELECT 'up SQL query';
CREATE TABLE invoice_credit_applications (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  invoice_id BIGINT NOT NULL,
  credit_note_id BIGINT NOT NULL,
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoice_credit_applications_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_credit_applications_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_credit_applications_credit_note_id
    FOREIGN KEY (credit_note_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_invoice_credit_applications_invoice ON invoice_credit_applications(invoice_id);

CREATE TABLE invoice_contra_settlements (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  customer_invoice_id BIGINT NOT NULL,
  supplier_bill_id BIGINT NOT NULL,
  amount NUMERIC(18,4),
  movement_id BIGINT,
  date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoice_contra_settlements_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_contra_settlements_customer_invoice_id
    FOREIGN KEY (customer_invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_contra_settlements_supplier_bill_id
    FOREIGN KEY (supplier_bill_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_invoice_contra_settlements_customer ON invoice_contra_settlements(customer_invoice_id);
CREATE INDEX idx_invoice_contra_settlements_vendor ON invoice_contra_settlements(supplier_bill_id);

CREATE TRIGGER enforce_ics_move BEFORE INSERT OR UPDATE OF movement_id ON invoice_contra_settlements
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_ics_move ON invoice_contra_settlements;
DROP INDEX IF EXISTS idx_invoice_contra_settlements_vendor;
DROP INDEX IF EXISTS idx_invoice_contra_settlements_customer;
DROP TABLE invoice_contra_settlements;
DROP INDEX IF EXISTS idx_invoice_credit_applications_invoice;
DROP TABLE invoice_credit_applications;
