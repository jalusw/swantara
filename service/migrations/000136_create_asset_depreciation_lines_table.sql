-- +goose Up
SELECT 'up SQL query';
CREATE TABLE asset_depreciation_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  asset_id BIGINT,
  sequence INT,
  depreciation_date DATE,
  amount NUMERIC(18,4),
  accumulated NUMERIC(18,4),
  remaining_value NUMERIC(18,4),
  movement_id BIGINT,
  posted BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_asset_depreciation_lines_asset_id
    FOREIGN KEY (asset_id)
    REFERENCES fixed_assets(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_adl_move BEFORE INSERT OR UPDATE OF movement_id ON asset_depreciation_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_adl_move ON asset_depreciation_lines;
DROP TABLE asset_depreciation_lines;
