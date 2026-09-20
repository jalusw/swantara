-- +goose Up
SELECT 'up SQL query';
CREATE TABLE commission_assignments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  plan_id BIGINT,
  salesperson_id BIGINT,
  date_start DATE,
  date_end DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_commission_assignments_plan_id
    FOREIGN KEY (plan_id)
    REFERENCES commission_plans(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_commission_assignments_salesperson_id
    FOREIGN KEY (salesperson_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE commission_assignments;
