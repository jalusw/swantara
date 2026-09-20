-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pos_payments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  order_id BIGINT,
  method TEXT,
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_payments_order_id
    FOREIGN KEY (order_id)
    REFERENCES pos_orders(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pos_payments;
