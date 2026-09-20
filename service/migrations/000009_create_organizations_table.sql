-- +goose Up
SELECT 'up SQL query';
CREATE TABLE organizations (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL,
  legal_name TEXT,
  parent_id BIGINT,
  base_currency CHAR(3) NOT NULL,
  country_code CHAR(2),
  tax_id TEXT,
  logo TEXT,
  timezone TEXT DEFAULT 'UTC',
  tax_year_start_month SMALLINT DEFAULT 1,
  auto_checkout_hour SMALLINT,
  rounding_minutes SMALLINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_organizations_parent_id
    FOREIGN KEY (parent_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE organizations;
