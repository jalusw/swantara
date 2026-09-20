-- +goose Up
SELECT 'up SQL query';
CREATE TABLE consumed_materials (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  production_order_id BIGINT NOT NULL,
  item_id BIGINT NOT NULL,
  qty_planned NUMERIC(18,4),
  qty_consumed NUMERIC(18,4) DEFAULT 0,
  unit_id BIGINT,
  stock_movement_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_consumed_materials_production_order_id
    FOREIGN KEY (production_order_id)
    REFERENCES production_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_consumed_materials_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_consumed_materials_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_moc_stock_movement BEFORE INSERT OR UPDATE OF stock_movement_id ON consumed_materials
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'stock_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_moc_stock_movement ON consumed_materials;
DROP TABLE consumed_materials;
