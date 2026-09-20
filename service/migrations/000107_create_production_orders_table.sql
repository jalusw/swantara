-- +goose Up
SELECT 'up SQL query';
CREATE TABLE production_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  item_id BIGINT NOT NULL,
  recipe_id BIGINT,
  qty_to_produce NUMERIC(18,4) NOT NULL,
  qty_produced NUMERIC(18,4) DEFAULT 0,
  unit_id BIGINT,
  src_location_id BIGINT,
  dst_location_id BIGINT,
  state TEXT CHECK (state IN ('draft','confirmed','planned','in_progress','done','cancelled')),
  date_planned_start TIMESTAMP,
  date_planned_finish TIMESTAMP,
  date_start TIMESTAMP,
  date_finished TIMESTAMP,
  origin TEXT,
  priority SMALLINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_production_orders_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_orders_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_orders_recipe_id
    FOREIGN KEY (recipe_id)
    REFERENCES recipes(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_orders_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_orders_src_location_id
    FOREIGN KEY (src_location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_production_orders_dst_location_id
    FOREIGN KEY (dst_location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE production_orders;
