-- +goose Up
SELECT 'up SQL query';
CREATE TABLE timesheets (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id BIGINT,
  date DATE,
  project_id BIGINT,
  task_id BIGINT,
  dimension_id BIGINT,
  hours NUMERIC(8,2),
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_timesheets_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_timesheets_project_id
    FOREIGN KEY (project_id)
    REFERENCES projects(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_timesheets_task_id
    FOREIGN KEY (task_id)
    REFERENCES project_tasks(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_timesheets_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE timesheets;
