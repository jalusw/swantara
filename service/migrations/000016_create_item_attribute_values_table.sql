-- +goose Up
SELECT 'up SQL query';
CREATE TABLE item_attribute_values (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  attribute_id BIGINT,
  value TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_item_attribute_values_attribute_id
    FOREIGN KEY (attribute_id)
    REFERENCES item_attributes(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE item_attribute_values;
