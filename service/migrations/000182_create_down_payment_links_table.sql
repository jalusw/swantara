-- +goose Up
SELECT 'up SQL query';
CREATE TABLE down_payment_links (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  advance_invoice_id BIGINT NOT NULL,
  final_invoice_id BIGINT NOT NULL,
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_down_payment_links_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_down_payment_links_advance_invoice_id
    FOREIGN KEY (advance_invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_down_payment_links_final_invoice_id
    FOREIGN KEY (final_invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_down_payment_links_final ON down_payment_links(final_invoice_id);

-- +goose Down
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_down_payment_links_final;
DROP TABLE down_payment_links;
