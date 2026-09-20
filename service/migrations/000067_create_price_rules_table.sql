-- +goose Up
SELECT 'up SQL query';
CREATE TABLE price_rules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  price_book_id BIGINT,
  applies_to TEXT CHECK (applies_to IN ('all','category','item','variant')),
  item_id BIGINT,
  category_id BIGINT,
  min_qty NUMERIC(18,4) DEFAULT 0,
  compute_type TEXT CHECK (compute_type IN ('fixed','percent','formula')),
  fixed_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4),
  date_start DATE,
  date_end DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_price_rules_price_book_id
    FOREIGN KEY (price_book_id)
    REFERENCES price_books(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_price_rules_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_price_rules_category_id
    FOREIGN KEY (category_id)
    REFERENCES item_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE price_rules;
