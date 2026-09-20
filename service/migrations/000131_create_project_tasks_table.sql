-- +goose Up
SELECT 'up SQL query';
CREATE TABLE project_tasks (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  project_id BIGINT,
  name TEXT,
  assignee_id BIGINT,
  stage TEXT,
  planned_hours NUMERIC(8,2),
  effective_hours NUMERIC(8,2),
  parent_task_id BIGINT,
  deadline DATE,
  priority SMALLINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_project_tasks_project_id
    FOREIGN KEY (project_id)
    REFERENCES projects(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_project_tasks_assignee_id
    FOREIGN KEY (assignee_id)
    REFERENCES users(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_project_tasks_parent_task_id
    FOREIGN KEY (parent_task_id)
    REFERENCES project_tasks(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE project_tasks;
