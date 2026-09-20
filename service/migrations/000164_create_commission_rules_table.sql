-- +goose Up
SELECT 'up SQL query';
CREATE TABLE commission_rules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  plan_id BIGINT,
  item_category_id BIGINT,
  min_amount NUMERIC(18,4),
  max_amount NUMERIC(18,4),
  rate_pct NUMERIC(8,4),
  fixed_amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_commission_rules_plan_id
    FOREIGN KEY (plan_id)
    REFERENCES commission_plans(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_commission_rules_item_category_id
    FOREIGN KEY (item_category_id)
    REFERENCES item_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE commission_rules;
