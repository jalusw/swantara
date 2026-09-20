-- +goose Up
SELECT 'up SQL query';
CREATE TABLE coupons (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  code TEXT,
  discount_type TEXT,
  discount_value NUMERIC(18,4),
  price_rule_id BIGINT,
  usage_limit INT,
  used_count INT DEFAULT 0,
  expiry_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_coupons_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_coupons_price_rule_id
    FOREIGN KEY (price_rule_id)
    REFERENCES price_rules(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE coupons;
