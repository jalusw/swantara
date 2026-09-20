-- +goose Up
SELECT 'up SQL query';
CREATE TABLE equipments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  item_id BIGINT,
  serial_batch_id BIGINT,
  owner_contact_id BIGINT,
  fixed_asset_id BIGINT,
  location TEXT,
  install_date DATE,
  warranty_end DATE,
  category TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_equipments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_equipments_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_equipments_serial_batch_id
    FOREIGN KEY (serial_batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_equipments_owner_contact_id
    FOREIGN KEY (owner_contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_equipments_fixed_asset_id
    FOREIGN KEY (fixed_asset_id)
    REFERENCES fixed_assets(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE equipments;
