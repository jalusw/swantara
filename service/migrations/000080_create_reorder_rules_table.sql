-- +goose Up
SELECT 'up SQL query';
CREATE TABLE reorder_rules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id BIGINT NOT NULL,
  warehouse_id BIGINT,
  location_id BIGINT,
  min_qty NUMERIC(18,4),
  max_qty NUMERIC(18,4),
  qty_multiple NUMERIC(18,4) DEFAULT 1,
  lead_time_days INT,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_reorder_rules_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_reorder_rules_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_reorder_rules_location_id
    FOREIGN KEY (location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE reorder_rules;
