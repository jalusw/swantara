-- +goose Up
SELECT 'up SQL query';
CREATE TABLE contact_suppliers (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  contact_id BIGINT NOT NULL UNIQUE,
  supplier_payment_term_id BIGINT,
  payable_account_id BIGINT,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_contact_suppliers_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_contact_suppliers_supplier_payment_term_id
    FOREIGN KEY (supplier_payment_term_id)
    REFERENCES payment_terms(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_contact_suppliers_payable_account_id
    FOREIGN KEY (payable_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE contact_suppliers;
