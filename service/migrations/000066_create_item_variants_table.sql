-- +goose Up
SELECT 'up SQL query';
CREATE TABLE item_variants (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id BIGINT NOT NULL,
  sku TEXT UNIQUE,
  barcode TEXT,
  attribute_json JSONB,
  extra_cost NUMERIC(18,4) DEFAULT 0,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_item_variants_item_id
    FOREIGN KEY (item_id)
    REFERENCES items(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE item_variants;
