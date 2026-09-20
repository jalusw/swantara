-- +goose Up
SELECT 'up SQL query';
CREATE TABLE work_centers (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT NOT NULL,
  code TEXT,
  cost_per_hour NUMERIC(18,4),
  capacity_per_hour NUMERIC(18,4),
  efficiency_pct NUMERIC(6,2) DEFAULT 100,
  oee_target NUMERIC(6,2),
  dimension_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_work_centers_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_work_centers_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE work_centers;
