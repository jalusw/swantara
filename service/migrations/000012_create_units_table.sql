-- +goose Up
SELECT 'up SQL query';
CREATE TABLE units (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  category_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  factor NUMERIC(28,8) NOT NULL,
  unit_type TEXT CHECK (unit_type IN ('reference','bigger','smaller')),
  rounding NUMERIC(12,6) DEFAULT 0.01,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_units_category_id
    FOREIGN KEY (category_id)
    REFERENCES unit_groups(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE units;
