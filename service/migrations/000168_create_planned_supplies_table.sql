-- +goose Up
SELECT 'up SQL query';
CREATE TABLE planned_supplies (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  planning_run_id BIGINT,
  item_id BIGINT,
  warehouse_id BIGINT,
  src_warehouse_id BIGINT REFERENCES warehouses (id),
  type TEXT CHECK (type IN ('purchase','manufacture','transfer')),
  qty NUMERIC(18,4),
  order_date DATE,
  due_date DATE,
  pegged_demand_id BIGINT,
  confirmed BOOLEAN DEFAULT false,
  generated_doc_type TEXT,
  generated_doc_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_planned_supplies_planning_run_id
    FOREIGN KEY (planning_run_id)
    REFERENCES planning_runs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_planned_supplies_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_planned_supplies_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_planned_supplies_pegged_demand_id
    FOREIGN KEY (pegged_demand_id)
    REFERENCES planning_needs(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE planned_supplies;
