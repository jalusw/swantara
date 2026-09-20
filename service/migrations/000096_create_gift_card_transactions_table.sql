-- +goose Up
SELECT 'up SQL query';
CREATE TABLE gift_card_transactions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  gift_card_id BIGINT,
  type TEXT CHECK (type IN ('issue','redeem','refund','adjust','forfeit')),
  amount NUMERIC(18,4),
  order_type TEXT,
  order_id BIGINT,
  movement_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_gift_card_transactions_gift_card_id
    FOREIGN KEY (gift_card_id)
    REFERENCES gift_cards(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_gct_move BEFORE INSERT OR UPDATE OF movement_id ON gift_card_transactions
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_gct_move ON gift_card_transactions;
DROP TABLE gift_card_transactions;
