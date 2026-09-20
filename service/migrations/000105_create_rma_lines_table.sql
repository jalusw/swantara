-- +goose Up
SELECT 'up SQL query';
CREATE TABLE rma_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  rma_id BIGINT,
  item_id BIGINT,
  qty NUMERIC(18,4),
  batch_id BIGINT,
  disposition TEXT CHECK (disposition IN ('restock','scrap','repair','replace')),
  stock_movement_id BIGINT,
  credit_note_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_rma_lines_rma_id
    FOREIGN KEY (rma_id)
    REFERENCES rmas(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rma_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rma_lines_batch_id
    FOREIGN KEY (batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rma_lines_credit_note_id
    FOREIGN KEY (credit_note_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_rma_stock_movement BEFORE INSERT OR UPDATE OF stock_movement_id ON rma_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'stock_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_rma_stock_movement ON rma_lines;
DROP TABLE rma_lines;
