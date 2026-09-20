-- +goose Up
SELECT 'up SQL query';
CREATE TABLE shop_tasks (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  production_order_id BIGINT NOT NULL,
  production_step_id BIGINT,
  work_center_id BIGINT NOT NULL,
  name TEXT,
  state TEXT CHECK (state IN ('draft','confirmed','planned','in_progress','done','cancelled')),
  sequence INT,
  planned_start TIMESTAMP,
  planned_finish TIMESTAMP,
  date_start TIMESTAMP,
  date_finished TIMESTAMP,
  planned_minutes NUMERIC(12,2) DEFAULT 0,
  actual_minutes NUMERIC(12,2) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_shop_tasks_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shop_tasks_production_order_id
    FOREIGN KEY (production_order_id)
    REFERENCES production_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shop_tasks_production_step_id
    FOREIGN KEY (production_step_id)
    REFERENCES production_steps(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shop_tasks_work_center_id
    FOREIGN KEY (work_center_id)
    REFERENCES work_centers(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE shop_tasks;