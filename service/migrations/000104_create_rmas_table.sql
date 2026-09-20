-- +goose Up
SELECT 'up SQL query';
CREATE TABLE rmas (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  type TEXT CHECK (type IN ('customer_return','vendor_return')),
  contact_id BIGINT,
  origin_order_type TEXT,
  origin_order_id BIGINT,
  reason TEXT,
  state TEXT CHECK (state IN ('draft','confirmed','received','refunded','done','cancelled')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_rmas_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rmas_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE rmas;
