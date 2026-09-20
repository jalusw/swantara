-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pos_sessions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  config_id BIGINT,
  cashier_id BIGINT,
  opened_at TIMESTAMP,
  closed_at TIMESTAMP,
  opening_balance NUMERIC(18,4),
  closing_balance NUMERIC(18,4),
  state TEXT CHECK (state IN ('opened','closing','closed')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_sessions_config_id
    FOREIGN KEY (config_id)
    REFERENCES pos_configs(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_sessions_cashier_id
    FOREIGN KEY (cashier_id)
    REFERENCES users(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pos_sessions_config_open
  ON pos_sessions (config_id)
  WHERE state = 'opened' AND deleted_at IS NULL;

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pos_sessions;
