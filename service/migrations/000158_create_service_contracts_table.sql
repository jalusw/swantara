-- +goose Up
SELECT 'up SQL query';
CREATE TABLE service_contracts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  contact_id BIGINT,
  equipment_id BIGINT,
  subscription_id BIGINT,
  coverage TEXT,
  sla_response_hours INT,
  date_start DATE,
  date_end DATE,
  state TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_service_contracts_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_contracts_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_contracts_equipment_id
    FOREIGN KEY (equipment_id)
    REFERENCES equipments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_contracts_subscription_id
    FOREIGN KEY (subscription_id)
    REFERENCES subscriptions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE service_contracts;
