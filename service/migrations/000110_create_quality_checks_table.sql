-- +goose Up
SELECT 'up SQL query';
CREATE TABLE quality_checks (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  point_id BIGINT,
  item_id BIGINT,
  batch_id BIGINT,
  shipment_id BIGINT,
  production_order_id BIGINT,
  measured_value NUMERIC(18,4),
  result TEXT CHECK (result IN ('pending','pass','fail')),
  checked_by BIGINT,
  checked_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_quality_checks_point_id
    FOREIGN KEY (point_id)
    REFERENCES quality_points(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_checks_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_checks_batch_id
    FOREIGN KEY (batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_checks_shipment_id
    FOREIGN KEY (shipment_id)
    REFERENCES shipments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_checks_production_order_id
    FOREIGN KEY (production_order_id)
    REFERENCES production_orders(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_checks_checked_by
    FOREIGN KEY (checked_by)
    REFERENCES users(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE quality_checks;
