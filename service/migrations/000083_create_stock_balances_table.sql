-- +goose Up
SELECT 'up SQL query';
CREATE TABLE stock_balances (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  item_id BIGINT NOT NULL,
  location_id BIGINT NOT NULL,
  batch_id BIGINT,
  quantity NUMERIC(18,4) DEFAULT 0,
  reserved_qty NUMERIC(18,4) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_stock_balances_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_balances_location_id
    FOREIGN KEY (location_id)
    REFERENCES stock_locations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_balances_batch_id
    FOREIGN KEY (batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_stock_balances_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  UNIQUE (item_id, location_id, batch_id)
);

CREATE INDEX idx_quant_lookup ON stock_balances(item_id, location_id);

-- +goose Down
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_quant_lookup;
DROP TABLE stock_balances;
