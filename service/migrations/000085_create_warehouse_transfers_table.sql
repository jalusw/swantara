-- +goose Up
SELECT 'up SQL query';
CREATE TABLE warehouse_transfers (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  src_warehouse_id BIGINT,
  dst_warehouse_id BIGINT,
  state TEXT CHECK (state IN ('draft','sent','in_transit','received','cancelled')),
  out_shipment_id BIGINT,
  in_shipment_id BIGINT,
  is_interorganization BOOLEAN DEFAULT false,
  scheduled_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_warehouse_transfers_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_warehouse_transfers_src_warehouse_id
    FOREIGN KEY (src_warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_warehouse_transfers_dst_warehouse_id
    FOREIGN KEY (dst_warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_warehouse_transfers_out_shipment_id
    FOREIGN KEY (out_shipment_id)
    REFERENCES shipments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_warehouse_transfers_in_shipment_id
    FOREIGN KEY (in_shipment_id)
    REFERENCES shipments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE warehouse_transfers;
