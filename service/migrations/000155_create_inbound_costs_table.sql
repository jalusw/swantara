-- +goose Up
SELECT 'up SQL query';
CREATE TABLE inbound_costs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  date DATE,
  state TEXT CHECK (state IN ('draft','posted','cancelled')),
  target_shipment_ids BIGINT[],
  movement_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_inbound_costs_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_lc_move BEFORE INSERT OR UPDATE OF movement_id ON inbound_costs
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_lc_move ON inbound_costs;
DROP TABLE inbound_costs;
