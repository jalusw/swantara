-- +goose Up
SELECT 'up SQL query';
CREATE TABLE shipments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  type TEXT CHECK (type IN ('incoming','outgoing','internal')),
  contact_id BIGINT,
  src_location_id BIGINT,
  dst_location_id BIGINT,
  state TEXT CHECK (state IN ('draft','waiting','confirmed','assigned','done','cancelled')),
  scheduled_date TIMESTAMP,
  date_done TIMESTAMP,
  origin TEXT,
  carrier_id BIGINT,
  tracking_ref TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_shipments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shipments_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shipments_src_location_id
    FOREIGN KEY (src_location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shipments_dst_location_id
    FOREIGN KEY (dst_location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_shipments_carrier_id
    FOREIGN KEY (carrier_id)
    REFERENCES carriers(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE shipments;
