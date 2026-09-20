-- +goose Up
SELECT 'up SQL query';
CREATE TABLE budget_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  budget_id BIGINT,
  account_id BIGINT,
  dimension_id BIGINT,
  planned_amount NUMERIC(18,4),
  practical_amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_budget_lines_budget_id
    FOREIGN KEY (budget_id)
    REFERENCES budgets(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_budget_lines_account_id
    FOREIGN KEY (account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_budget_lines_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE budget_lines;
