-- +goose Up
SELECT 'up SQL query';
CREATE TABLE subscription_plans (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  recurring_interval TEXT,
  recurring_count INT DEFAULT 1,
  organization_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_subscription_plans_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_subscription_plans_organization_id ON subscription_plans(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE subscription_plans;
