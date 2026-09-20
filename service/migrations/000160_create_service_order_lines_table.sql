-- +goose Up
SELECT 'up SQL query';
CREATE TABLE service_order_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  service_order_id BIGINT,
  type TEXT CHECK (type IN ('part','labor','expense')),
  item_id BIGINT,
  description TEXT,
  qty NUMERIC(18,4),
  unit_id BIGINT,
  unit_cost NUMERIC(18,4),
  unit_price NUMERIC(18,4),
  stock_movement_id BIGINT,
  billable BOOLEAN DEFAULT true,
  covered_by_warranty BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_service_order_lines_service_order_id
    FOREIGN KEY (service_order_id)
    REFERENCES service_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_order_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_order_lines_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_sol_stock_movement BEFORE INSERT OR UPDATE OF stock_movement_id ON service_order_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'stock_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_sol_stock_movement ON service_order_lines;
DROP TABLE service_order_lines;
