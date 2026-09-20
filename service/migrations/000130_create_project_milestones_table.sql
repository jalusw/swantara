-- +goose Up
SELECT 'up SQL query';
CREATE TABLE project_milestones (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  project_id BIGINT,
  name TEXT,
  deadline DATE,
  reached BOOLEAN DEFAULT false,
  sale_line_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_project_milestones_project_id
    FOREIGN KEY (project_id)
    REFERENCES projects(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_project_milestones_sale_line_id
    FOREIGN KEY (sale_line_id)
    REFERENCES sale_order_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE project_milestones;
