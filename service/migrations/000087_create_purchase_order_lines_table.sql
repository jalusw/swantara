-- +goose Up
SELECT 'up SQL query';
CREATE TABLE purchase_order_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  order_id BIGINT NOT NULL,
  sequence INT DEFAULT 10,
  item_id BIGINT,
  description TEXT,
  qty_ordered NUMERIC(18,4) NOT NULL,
  qty_received NUMERIC(18,4) DEFAULT 0,
  qty_billed NUMERIC(18,4) DEFAULT 0,
  qty_returns NUMERIC(18,4) DEFAULT 0,
  unit_id BIGINT,
  unit_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4) DEFAULT 0,
  tax_ids BIGINT[],
  dimension_id BIGINT,
  price_subtotal NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_purchase_order_lines_order_id
    FOREIGN KEY (order_id)
    REFERENCES purchase_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_order_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_order_lines_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_order_lines_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_poline_order ON purchase_order_lines(order_id);

-- +goose Down
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_poline_order;
DROP TABLE purchase_order_lines;
