-- +goose Up
SELECT 'up SQL query';
CREATE TABLE stock_count_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  stock_count_id BIGINT,
  item_id BIGINT,
  batch_id BIGINT,
  theoretical_qty NUMERIC(18,4),
  counted_qty NUMERIC(18,4),
  diff_qty NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_stock_count_lines_stock_count_id
    FOREIGN KEY (stock_count_id)
    REFERENCES stock_counts(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_count_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_count_lines_batch_id
    FOREIGN KEY (batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE stock_count_lines;
