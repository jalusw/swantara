-- +goose Up
SELECT 'up SQL query';
CREATE TABLE users(
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  username TEXT NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT,
  email TEXT NOT NULL,
  phone TEXT,
  password TEXT NOT NULL,
  avatar TEXT,
  bio TEXT,
  birthday DATE DEFAULT NULL,
  active BOOLEAN DEFAULT FALSE,
  private BOOLEAN DEFAULT FALSE,
  sex TEXT CHECK (sex IN ('male', 'female')),
  address TEXT,
  city TEXT,
  postal_code TEXT,
  email_verified_at TIMESTAMP DEFAULT NULL,
  phone_verified_at TIMESTAMP DEFAULT NULL,
  contact_id BIGINT,
  last_login TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_username_unique_active ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_email_unique_active ON users(email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_phone_unique ON users(phone) WHERE deleted_at IS NULL;

-- +goose Down
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_users_username_unique_active;
DROP INDEX IF EXISTS idx_users_email_unique_active;
DROP INDEX IF EXISTS idx_users_phone_unique;
DROP TABLE users;
