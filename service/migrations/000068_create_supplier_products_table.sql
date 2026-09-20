-- +goose Up
SELECT 'up SQL query';
CREATE TABLE supplier_products (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id BIGINT NOT NULL,
  supplier_id BIGINT NOT NULL,
  supplier_sku TEXT,
  supplier_product_name TEXT,
  min_qty NUMERIC(18,4) DEFAULT 1,
  price NUMERIC(18,4),
  currency_code CHAR(3),
  lead_time_days INT,
  priority INT DEFAULT 10,
  valid_from DATE,
  valid_to DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_supplier_products_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_products_supplier_id
    FOREIGN KEY (supplier_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_products_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE supplier_products;
