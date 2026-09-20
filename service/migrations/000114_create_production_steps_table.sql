-- +goose Up
SELECT 'up SQL query';
CREATE TABLE production_steps (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  recipe_id BIGINT,
  work_center_id BIGINT,
  name TEXT,
  sequence INT,
  setup_minutes NUMERIC(12,2) DEFAULT 0,
  time_minutes NUMERIC(12,2),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_production_steps_recipe_id
    FOREIGN KEY (recipe_id)
    REFERENCES recipes(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_steps_work_center_id
    FOREIGN KEY (work_center_id)
    REFERENCES work_centers(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE production_steps;
