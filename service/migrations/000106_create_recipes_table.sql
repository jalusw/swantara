-- +goose Up
SELECT 'up SQL query';
CREATE TABLE recipes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  item_id BIGINT NOT NULL,
  code TEXT,
  qty NUMERIC(18,4) DEFAULT 1,
  unit_id BIGINT,
  type TEXT CHECK (type IN ('manufacture','kit','subcontract')),
  version INT DEFAULT 1,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_boms_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_boms_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_boms_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE recipes;
