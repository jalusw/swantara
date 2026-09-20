-- +goose Up
SELECT 'up SQL query';
CREATE TABLE service_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  contact_id BIGINT,
  equipment_id BIGINT,
  contract_id BIGINT,
  type TEXT CHECK (type IN ('repair','maintenance','installation','inspection')),
  priority SMALLINT,
  state TEXT CHECK (state IN ('new','scheduled','in_progress','done','invoiced','cancelled')),
  scheduled_date TIMESTAMP,
  technician_id BIGINT,
  invoice_id BIGINT,
  dimension_id BIGINT,
  reported_issue TEXT,
  resolution TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_service_orders_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_equipment_id
    FOREIGN KEY (equipment_id)
    REFERENCES equipments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_contract_id
    FOREIGN KEY (contract_id)
    REFERENCES service_contracts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_technician_id
    FOREIGN KEY (technician_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_service_orders_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE service_orders;
