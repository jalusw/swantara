-- +goose Up
SELECT 'up SQL query';
CREATE TABLE carriers (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  tracking_url_tpl TEXT,
  delivery_item_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE carriers;
