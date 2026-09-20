-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payment_allocations (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  payment_id BIGINT,
  invoice_id BIGINT,
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payment_allocations_payment_id
    FOREIGN KEY (payment_id)
    REFERENCES payments(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_payment_allocations_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE payment_allocations;
