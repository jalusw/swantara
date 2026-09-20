-- +goose Up
SELECT 'up SQL query';
CREATE TABLE asset_categories (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  organization_id BIGINT,
  asset_account_id BIGINT,
  depreciation_account_id BIGINT,
  expense_account_id BIGINT,
  method TEXT CHECK (method IN ('linear','declining','declining_then_linear')),
  method_number INT,
  method_period TEXT,
  gain_account_id BIGINT,
  loss_account_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_asset_categories_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_asset_categories_asset_account_id
    FOREIGN KEY (asset_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_asset_categories_depreciation_account_id
    FOREIGN KEY (depreciation_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_asset_categories_expense_account_id
    FOREIGN KEY (expense_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_asset_categories_gain_account_id
    FOREIGN KEY (gain_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_asset_categories_loss_account_id
    FOREIGN KEY (loss_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE asset_categories;
