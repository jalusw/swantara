-- +goose Up
CREATE TABLE invoice_installments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  invoice_id BIGINT NOT NULL,
  sequence INT,
  due_date DATE,
  amount NUMERIC(18,4),
  state TEXT CHECK (state IN ('pending','paid','cancelled')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoice_installments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_installments_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

CREATE INDEX idx_invoice_installments_invoice ON invoice_installments(invoice_id);

-- +goose Down
DROP INDEX IF EXISTS idx_invoice_installments_invoice;
DROP TABLE invoice_installments;
