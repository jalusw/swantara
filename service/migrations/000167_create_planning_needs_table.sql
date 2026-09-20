-- +goose Up
SELECT 'up SQL query';
CREATE TABLE planning_needs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  planning_run_id BIGINT,
  item_id BIGINT,
  warehouse_id BIGINT,
  source_type TEXT,
  source_id BIGINT,
  qty NUMERIC(18,4),
  required_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_planning_needs_planning_run_id
    FOREIGN KEY (planning_run_id)
    REFERENCES planning_runs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_planning_needs_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_planning_needs_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE planning_needs;
