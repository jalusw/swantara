-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payment_terms (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL,
  note TEXT,
  organization_id BIGINT NOT NULL,
  code TEXT,
  is_active BOOLEAN DEFAULT true,
  template_key TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payment_terms_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_payment_terms_organization_id ON payment_terms(organization_id);
CREATE UNIQUE INDEX idx_payment_terms_org_name ON payment_terms(organization_id, name) WHERE deleted_at IS NULL;

-- +goose Down
SELECT 'down SQL query';
DROP TABLE payment_terms;
