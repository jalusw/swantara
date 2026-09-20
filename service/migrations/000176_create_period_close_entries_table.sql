-- +goose Up
SELECT 'up SQL query';
CREATE TABLE period_close_entries (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  period_id BIGINT NOT NULL,
  movement_id BIGINT,
  closing_date DATE NOT NULL,
  state TEXT CHECK (state IN ('draft','posted')),
  posted_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_period_close_entries_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  CONSTRAINT fk_period_close_entries_period_id FOREIGN KEY (period_id) REFERENCES tax_periods(id) ON DELETE RESTRICT
);
CREATE INDEX idx_period_close_entries_period_id ON period_close_entries(period_id);
CREATE INDEX idx_period_close_entries_organization_id ON period_close_entries(organization_id);
CREATE TRIGGER enforce_pce_move BEFORE INSERT OR UPDATE OF movement_id ON period_close_entries
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_pce_move ON period_close_entries;
DROP TABLE period_close_entries;
