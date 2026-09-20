-- +goose Up
SELECT 'up SQL query';
CREATE TABLE fixed_assets (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  category_id BIGINT,
  purchase_value NUMERIC(18,4),
  salvage_value NUMERIC(18,4) DEFAULT 0,
  acquisition_date DATE,
  in_service_date DATE,
  original_movement_id BIGINT,
  invoice_line_id BIGINT,
  state TEXT CHECK (state IN ('draft','running','disposed','sold')),
  disposal_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_fixed_assets_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_fixed_assets_category_id
    FOREIGN KEY (category_id)
    REFERENCES asset_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_fixed_assets_invoice_line_id
    FOREIGN KEY (invoice_line_id)
    REFERENCES invoice_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_fa_original BEFORE INSERT OR UPDATE OF original_movement_id ON fixed_assets
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'original_movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_fa_original ON fixed_assets;
DROP TABLE fixed_assets;
