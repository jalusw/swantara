-- +goose Up
SELECT 'up SQL query';
CREATE TABLE maintenance_plans (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  equipment_id BIGINT,
  name TEXT,
  interval_days INT,
  next_due DATE,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_maintenance_plans_equipment_id
    FOREIGN KEY (equipment_id)
    REFERENCES equipments(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE maintenance_plans;
