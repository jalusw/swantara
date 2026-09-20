-- +goose Up
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS dimension_distributions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  journal_line_id BIGINT NOT NULL,
  dimension_id BIGINT NOT NULL REFERENCES dimensions(id) ON DELETE RESTRICT,
  percent NUMERIC(5,2) NOT NULL,
  amount NUMERIC(18,4) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_dimension_dist_line ON dimension_distributions(journal_line_id);
CREATE INDEX IF NOT EXISTS idx_dimension_dist_dimension ON dimension_distributions(dimension_id);
CREATE TRIGGER enforce_ad_journal_line BEFORE INSERT OR UPDATE OF journal_line_id ON dimension_distributions
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_lines', 'journal_line_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_ad_journal_line ON dimension_distributions;
DROP TABLE IF EXISTS dimension_distributions;
