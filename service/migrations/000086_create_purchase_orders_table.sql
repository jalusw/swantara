-- +goose Up
SELECT 'up SQL query';
CREATE TABLE purchase_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  supplier_id BIGINT NOT NULL,
  supplier_ref TEXT,
  currency_code CHAR(3),
  warehouse_id BIGINT,
  dest_location_id BIGINT,
  state TEXT CHECK (state IN ('draft','sent','confirmed','done','cancelled')),
  order_date DATE,
  expected_date DATE,
  payment_term_id BIGINT,
  incoterm TEXT,
  amount_untaxed NUMERIC(18,4) DEFAULT 0,
  amount_tax NUMERIC(18,4) DEFAULT 0,
  amount_total NUMERIC(18,4) DEFAULT 0,
  invoice_status TEXT CHECK (invoice_status IN ('no','to_invoice','invoiced')),
  receipt_status TEXT CHECK (receipt_status IN ('pending','partial','done')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_purchase_orders_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_orders_supplier_id
    FOREIGN KEY (supplier_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_orders_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_orders_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_orders_dest_location_id
    FOREIGN KEY (dest_location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_orders_payment_term_id
    FOREIGN KEY (payment_term_id)
    REFERENCES payment_terms(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE purchase_orders;
