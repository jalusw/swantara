-- +goose Up
SELECT 'up SQL query';
CREATE TABLE recipe_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  recipe_id BIGINT NOT NULL,
  component_id BIGINT NOT NULL,
  qty NUMERIC(18,4) NOT NULL,
  unit_id BIGINT,
  scrap_pct NUMERIC(8,4) DEFAULT 0,
  production_step_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_recipe_lines_recipe_id
    FOREIGN KEY (recipe_id)
    REFERENCES recipes(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_recipe_lines_component_id
    FOREIGN KEY (component_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_recipe_lines_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_recipe_lines_production_step_id
    FOREIGN KEY (production_step_id)
    REFERENCES production_steps(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE recipe_lines;
