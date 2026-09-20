-- +goose Up
SELECT 'up SQL query';
CREATE TABLE item_categories (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL,
  parent_id BIGINT,
  income_account_id BIGINT,
  expense_account_id BIGINT,
  stock_cost_account_id BIGINT,
  stock_input_account_id BIGINT,
  stock_output_account_id BIGINT,
  cogs_account_id BIGINT,
  cost_method TEXT CHECK (cost_method IN ('standard','fifo','average')),
  valuation TEXT CHECK (valuation IN ('manual','automated')),
  organization_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_item_categories_parent_id
    FOREIGN KEY (parent_id)
    REFERENCES item_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_income_account_id
    FOREIGN KEY (income_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_expense_account_id
    FOREIGN KEY (expense_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_stock_cost_account_id
    FOREIGN KEY (stock_cost_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_stock_input_account_id
    FOREIGN KEY (stock_input_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_stock_output_account_id
    FOREIGN KEY (stock_output_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_cogs_account_id
    FOREIGN KEY (cogs_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_item_categories_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_item_categories_organization_id ON item_categories(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE item_categories;
