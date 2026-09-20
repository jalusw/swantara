-- +goose Up
SELECT 'up SQL query';
CREATE TABLE organization_modules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  module_id TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_organization_modules_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  UNIQUE (organization_id, module_id)
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE organization_modules;
