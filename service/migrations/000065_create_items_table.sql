-- +goose Up
SELECT 'up SQL query';
CREATE TABLE items (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT NOT NULL,
  category_id BIGINT,
  type TEXT CHECK (type IN ('stockable','consumable','service','digital')),
  unit_id BIGINT,
  purchase_unit_id BIGINT,
  list_price NUMERIC(18,4),
  standard_cost NUMERIC(18,4),
  is_purchasable BOOLEAN DEFAULT true,
  is_sellable BOOLEAN DEFAULT true,
  is_manufactured BOOLEAN DEFAULT false,
  tracking TEXT CHECK (tracking IN ('none','lot','serial')),
  weight NUMERIC(18,4),
  volume NUMERIC(18,4),
  hs_code TEXT,
  description_sale TEXT,
  description_purchase TEXT,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_items_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_items_category_id
    FOREIGN KEY (category_id)
    REFERENCES item_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_items_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_items_purchase_unit_id
    FOREIGN KEY (purchase_unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE items;
