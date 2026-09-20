-- +goose Up
SELECT 'up SQL query';
CREATE TABLE cost_layers (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  movement_id BIGINT,
  item_id BIGINT NOT NULL,
  quantity NUMERIC(18,4),
  unit_cost NUMERIC(18,4),
  value NUMERIC(18,4),
  remaining_qty NUMERIC(18,4),
  remaining_value NUMERIC(18,4),
  journal_entry_id BIGINT,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_cost_layers_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_svl_journal_entry BEFORE INSERT OR UPDATE OF journal_entry_id ON cost_layers
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'journal_entry_id');
CREATE TRIGGER enforce_svl_move BEFORE INSERT OR UPDATE OF movement_id ON cost_layers
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'movement_id');

CREATE INDEX idx_val_open ON cost_layers(item_id) WHERE remaining_qty <> 0;

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_svl_journal_entry ON cost_layers;
DROP TRIGGER IF EXISTS enforce_svl_move ON cost_layers;
DROP INDEX IF EXISTS idx_val_open;
DROP TABLE cost_layers;
