-- +goose Up
SELECT 'up SQL query';
CREATE TABLE user_sessions(
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  device_name TEXT,
  os TEXT,
  browser TEXT,
  user_agent TEXT,
  refresh_token TEXT NOT NULL,

  ip_address TEXT,
  country TEXT,
  city TEXT,

  expires_at TIMESTAMP,
  revoked_at TIMESTAMP,
  last_used_at TIMESTAMP,

  user_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,

  CONSTRAINT fk_session_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE user_sessions;