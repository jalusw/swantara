-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pos_order_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  order_id BIGINT,
  item_id BIGINT,
  qty NUMERIC(18,4),
  unit_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4),
  tax_ids BIGINT[],
  price_subtotal NUMERIC(18,4),
  price_tax NUMERIC(18,4) DEFAULT 0,
  price_total NUMERIC(18,4) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_order_lines_order_id
    FOREIGN KEY (order_id)
    REFERENCES pos_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_order_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pos_order_lines;
