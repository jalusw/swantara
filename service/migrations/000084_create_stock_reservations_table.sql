-- +goose Up
SELECT 'up SQL query';
CREATE TABLE stock_holds (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  movement_id BIGINT,
  balance_id BIGINT,
  qty NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_stock_holds_balance_id
    FOREIGN KEY (balance_id)
    REFERENCES stock_balances(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_sr_move BEFORE INSERT OR UPDATE OF movement_id ON stock_holds
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_sr_move ON stock_holds;
DROP TABLE stock_holds;
