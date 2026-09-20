-- +goose Up
SELECT 'up SQL query';
CREATE TABLE dropship_links (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  sale_order_line_id BIGINT,
  purchase_order_line_id BIGINT,
  stock_movement_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_dropship_links_sale_order_line_id
    FOREIGN KEY (sale_order_line_id)
    REFERENCES sale_order_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_dropship_links_purchase_order_line_id
    FOREIGN KEY (purchase_order_line_id)
    REFERENCES purchase_order_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_dl_stock_movement BEFORE INSERT OR UPDATE OF stock_movement_id ON dropship_links
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('stock_movements', 'stock_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_dl_stock_movement ON dropship_links;
DROP TABLE dropship_links;
