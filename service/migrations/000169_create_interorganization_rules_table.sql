-- +goose Up
SELECT 'up SQL query';
CREATE TABLE interorganization_rules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  from_organization_id BIGINT,
  to_organization_id BIGINT,
  auto_mirror BOOLEAN DEFAULT true,
  supplier_contact_id BIGINT,
  customer_contact_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_interorganization_rules_from_organization_id
    FOREIGN KEY (from_organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_interorganization_rules_to_organization_id
    FOREIGN KEY (to_organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_interorganization_rules_supplier_contact_id
    FOREIGN KEY (supplier_contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_interorganization_rules_customer_contact_id
    FOREIGN KEY (customer_contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE interorganization_rules;