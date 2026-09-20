-- +goose Up
SELECT 'up SQL query';
CREATE TABLE inbound_cost_adjustments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  inbound_cost_id BIGINT,
  stock_movement_id BIGINT,
  item_id BIGINT,
  additional_cost NUMERIC(18,4),
  cost_layer_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_inbound_cost_adjustments_inbound_cost_id
    FOREIGN KEY (inbound_cost_id)
    REFERENCES inbound_costs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_inbound_cost_adjustments_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_inbound_cost_adjustments_cost_layer_id
    FOREIGN KEY (cost_layer_id)
    REFERENCES cost_layers(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_lca_stock_movement BEFORE INSERT OR UPDATE OF stock_movement_id ON inbound_cost_adjustments
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'stock_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_lca_stock_movement ON inbound_cost_adjustments;
DROP TABLE inbound_cost_adjustments;
