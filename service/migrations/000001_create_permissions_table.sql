-- +goose Up
SELECT 'up SQL query';
CREATE TABLE permissions(
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  description TEXT,
  action TEXT NOT NULL CHECK(action IN ('view','create','update','delete','approve','post')),
  resource TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_permissions_code
ON permissions(code);


-- +goose Down
SELECT 'down SQL query';
DROP TABLE permissions;