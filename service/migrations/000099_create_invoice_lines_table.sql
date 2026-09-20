-- +goose Up
SELECT 'up SQL query';
CREATE TABLE invoice_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  invoice_id BIGINT NOT NULL,
  sequence INT DEFAULT 10,
  item_id BIGINT,
  description TEXT,
  qty NUMERIC(18,4),
  unit_id BIGINT,
  unit_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4) DEFAULT 0,
  tax_ids BIGINT[],
  account_id BIGINT,
  dimension_id BIGINT,
  price_subtotal NUMERIC(18,4),
  sale_line_id BIGINT,
  purchase_line_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_invoice_lines_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_account_id
    FOREIGN KEY (account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_sale_line_id
    FOREIGN KEY (sale_line_id)
    REFERENCES sale_order_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_invoice_lines_purchase_line_id
    FOREIGN KEY (purchase_line_id)
    REFERENCES purchase_order_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE invoice_lines;
