-- +goose Up
SELECT 'up SQL query';
ALTER TABLE item_categories RENAME COLUMN stock_cost_account_id TO stock_valuation_account_id;
ALTER TABLE item_categories RENAME CONSTRAINT fk_item_categories_stock_cost_account_id TO fk_item_categories_stock_valuation_account_id;

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE item_categories RENAME CONSTRAINT fk_item_categories_stock_valuation_account_id TO fk_item_categories_stock_cost_account_id;
ALTER TABLE item_categories RENAME COLUMN stock_valuation_account_id TO stock_cost_account_id;
