-- +goose Up
SELECT 'up SQL query';
CREATE TABLE stock_locations (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  warehouse_id BIGINT,
  name TEXT NOT NULL,
  code TEXT,
  parent_id BIGINT,
  usage TEXT CHECK (usage IN ('internal','customer','supplier','production','inventory','transit','scrap','view')),
  barcode TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_stock_locations_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_locations_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_locations_parent_id
    FOREIGN KEY (parent_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE stock_locations;
