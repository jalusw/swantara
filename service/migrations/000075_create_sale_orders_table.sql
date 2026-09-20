-- +goose Up
SELECT 'up SQL query';
CREATE TABLE sale_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  contact_id BIGINT NOT NULL,
  ship_address_id BIGINT,
  bill_address_id BIGINT,
  price_book_id BIGINT,
  currency_code CHAR(3),
  salesperson_id BIGINT,
  sales_group_id BIGINT,
  prospect_id BIGINT,
  warehouse_id BIGINT,
  state TEXT CHECK (state IN ('draft','sent','confirmed','done','cancelled')),
  order_date DATE,
  expected_date DATE,
  validity_date DATE,
  payment_term_id BIGINT,
  incoterm TEXT,
  customer_po_ref TEXT,
  amount_untaxed NUMERIC(18,4) DEFAULT 0,
  amount_tax NUMERIC(18,4) DEFAULT 0,
  amount_total NUMERIC(18,4) DEFAULT 0,
  invoice_status TEXT CHECK (invoice_status IN ('no','to_invoice','invoiced')),
  delivery_status TEXT CHECK (delivery_status IN ('pending','partial','done')),
  note TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_sale_orders_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_ship_address_id
    FOREIGN KEY (ship_address_id)
    REFERENCES contact_addresses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_bill_address_id
    FOREIGN KEY (bill_address_id)
    REFERENCES contact_addresses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_price_book_id
    FOREIGN KEY (price_book_id)
    REFERENCES price_books(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_salesperson_id
    FOREIGN KEY (salesperson_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_sales_group_id
    FOREIGN KEY (sales_group_id)
    REFERENCES sales_groups(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_prospect_id
    FOREIGN KEY (prospect_id)
    REFERENCES prospects(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_sale_orders_payment_term_id
    FOREIGN KEY (payment_term_id)
    REFERENCES payment_terms(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE sale_orders;
