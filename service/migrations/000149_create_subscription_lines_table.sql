-- +goose Up
SELECT 'up SQL query';
CREATE TABLE subscription_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  subscription_id BIGINT,
  item_id BIGINT,
  qty NUMERIC(18,4),
  unit_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_subscription_lines_subscription_id
    FOREIGN KEY (subscription_id)
    REFERENCES subscriptions(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_subscription_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE subscription_lines;
