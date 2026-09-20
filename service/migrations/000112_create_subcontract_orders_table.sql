-- +goose Up
SELECT 'up SQL query';
CREATE TABLE outside_processing_orders (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  production_order_id BIGINT,
  supplier_id BIGINT,
  purchase_order_id BIGINT,
  state TEXT NOT NULL DEFAULT 'draft',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_outside_processing_orders_production_order_id
    FOREIGN KEY (production_order_id)
    REFERENCES production_orders(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_outside_processing_orders_supplier_id
    FOREIGN KEY (supplier_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_outside_processing_orders_purchase_order_id
    FOREIGN KEY (purchase_order_id)
    REFERENCES purchase_orders(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE outside_processing_orders;
