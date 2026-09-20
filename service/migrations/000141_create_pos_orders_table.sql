-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pos_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  session_id BIGINT,
  contact_id BIGINT,
  name TEXT,
  amount_total NUMERIC(18,4),
  amount_tax NUMERIC(18,4),
  state TEXT,
  invoice_id BIGINT,
  order_time TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_orders_session_id
    FOREIGN KEY (session_id)
    REFERENCES pos_sessions(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_orders_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_orders_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pos_orders;
