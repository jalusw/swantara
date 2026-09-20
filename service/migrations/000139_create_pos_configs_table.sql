-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pos_configs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  warehouse_id BIGINT,
  journal_id BIGINT,
  price_book_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pos_configs_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_configs_warehouse_id
    FOREIGN KEY (warehouse_id)
    REFERENCES warehouses(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_configs_journal_id
    FOREIGN KEY (journal_id)
    REFERENCES journals(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_pos_configs_price_book_id
    FOREIGN KEY (price_book_id)
    REFERENCES price_books(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pos_configs;
