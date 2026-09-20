-- +goose Up
SELECT 'up SQL query';
CREATE TABLE inbound_cost_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  inbound_cost_id BIGINT,
  item_id BIGINT,
  description TEXT,
  amount NUMERIC(18,4),
  supplier_bill_line_id BIGINT,
  split_method TEXT CHECK (split_method IN ('by_quantity','by_weight','by_volume','by_value','equal')),
  account_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_inbound_cost_lines_inbound_cost_id
    FOREIGN KEY (inbound_cost_id)
    REFERENCES inbound_costs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_inbound_cost_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_inbound_cost_lines_supplier_bill_line_id
    FOREIGN KEY (supplier_bill_line_id)
    REFERENCES invoice_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_inbound_cost_lines_account_id
    FOREIGN KEY (account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE inbound_cost_lines;
