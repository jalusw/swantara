-- +goose Up
SELECT 'up SQL query';
CREATE TABLE journal_entrys (
  id BIGINT GENERATED ALWAYS AS IDENTITY,
  organization_id BIGINT NOT NULL,
  journal_id BIGINT NOT NULL,
  name TEXT,
  date DATE NOT NULL,
  ref TEXT,
  state TEXT CHECK (state IN ('draft','posted','cancelled')),
  currency_code CHAR(3),
  origin_type TEXT,
  origin_id BIGINT,
  reversed_entry_id BIGINT,
  posted_at TIMESTAMP,
  posted_by BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (date, id),
  CONSTRAINT journal_entrys_origin_consistency
    CHECK (NOT (origin_id IS NOT NULL AND origin_type IS NULL)),
  CONSTRAINT chk_am_name_not_null
    CHECK (name IS NOT NULL OR state = 'draft'),
  CONSTRAINT fk_journal_entrys_organization_id
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_entrys_journal_id
    FOREIGN KEY (journal_id) REFERENCES journals(id) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_entrys_currency_code
    FOREIGN KEY (currency_code) REFERENCES currencies(code) ON DELETE RESTRICT,
  CONSTRAINT fk_journal_entrys_posted_by
    FOREIGN KEY (posted_by) REFERENCES users(id) ON DELETE RESTRICT
) PARTITION BY RANGE (date);

CREATE TABLE journal_entrys_default PARTITION OF journal_entrys DEFAULT;
CREATE TABLE journal_entrys_202001 PARTITION OF journal_entrys FOR VALUES FROM ('2020-01-01') TO ('2020-02-01');
CREATE TABLE journal_entrys_202002 PARTITION OF journal_entrys FOR VALUES FROM ('2020-02-01') TO ('2020-03-01');
CREATE TABLE journal_entrys_202003 PARTITION OF journal_entrys FOR VALUES FROM ('2020-03-01') TO ('2020-04-01');
CREATE TABLE journal_entrys_202004 PARTITION OF journal_entrys FOR VALUES FROM ('2020-04-01') TO ('2020-05-01');
CREATE TABLE journal_entrys_202005 PARTITION OF journal_entrys FOR VALUES FROM ('2020-05-01') TO ('2020-06-01');
CREATE TABLE journal_entrys_202006 PARTITION OF journal_entrys FOR VALUES FROM ('2020-06-01') TO ('2020-07-01');
CREATE TABLE journal_entrys_202007 PARTITION OF journal_entrys FOR VALUES FROM ('2020-07-01') TO ('2020-08-01');
CREATE TABLE journal_entrys_202008 PARTITION OF journal_entrys FOR VALUES FROM ('2020-08-01') TO ('2020-09-01');
CREATE TABLE journal_entrys_202009 PARTITION OF journal_entrys FOR VALUES FROM ('2020-09-01') TO ('2020-10-01');
CREATE TABLE journal_entrys_202010 PARTITION OF journal_entrys FOR VALUES FROM ('2020-10-01') TO ('2020-11-01');
CREATE TABLE journal_entrys_202011 PARTITION OF journal_entrys FOR VALUES FROM ('2020-11-01') TO ('2020-12-01');
CREATE TABLE journal_entrys_202012 PARTITION OF journal_entrys FOR VALUES FROM ('2020-12-01') TO ('2021-01-01');
CREATE TABLE journal_entrys_202101 PARTITION OF journal_entrys FOR VALUES FROM ('2021-01-01') TO ('2021-02-01');
CREATE TABLE journal_entrys_202102 PARTITION OF journal_entrys FOR VALUES FROM ('2021-02-01') TO ('2021-03-01');
CREATE TABLE journal_entrys_202103 PARTITION OF journal_entrys FOR VALUES FROM ('2021-03-01') TO ('2021-04-01');
CREATE TABLE journal_entrys_202104 PARTITION OF journal_entrys FOR VALUES FROM ('2021-04-01') TO ('2021-05-01');
CREATE TABLE journal_entrys_202105 PARTITION OF journal_entrys FOR VALUES FROM ('2021-05-01') TO ('2021-06-01');
CREATE TABLE journal_entrys_202106 PARTITION OF journal_entrys FOR VALUES FROM ('2021-06-01') TO ('2021-07-01');
CREATE TABLE journal_entrys_202107 PARTITION OF journal_entrys FOR VALUES FROM ('2021-07-01') TO ('2021-08-01');
CREATE TABLE journal_entrys_202108 PARTITION OF journal_entrys FOR VALUES FROM ('2021-08-01') TO ('2021-09-01');
CREATE TABLE journal_entrys_202109 PARTITION OF journal_entrys FOR VALUES FROM ('2021-09-01') TO ('2021-10-01');
CREATE TABLE journal_entrys_202110 PARTITION OF journal_entrys FOR VALUES FROM ('2021-10-01') TO ('2021-11-01');
CREATE TABLE journal_entrys_202111 PARTITION OF journal_entrys FOR VALUES FROM ('2021-11-01') TO ('2021-12-01');
CREATE TABLE journal_entrys_202112 PARTITION OF journal_entrys FOR VALUES FROM ('2021-12-01') TO ('2022-01-01');
CREATE TABLE journal_entrys_202201 PARTITION OF journal_entrys FOR VALUES FROM ('2022-01-01') TO ('2022-02-01');
CREATE TABLE journal_entrys_202202 PARTITION OF journal_entrys FOR VALUES FROM ('2022-02-01') TO ('2022-03-01');
CREATE TABLE journal_entrys_202203 PARTITION OF journal_entrys FOR VALUES FROM ('2022-03-01') TO ('2022-04-01');
CREATE TABLE journal_entrys_202204 PARTITION OF journal_entrys FOR VALUES FROM ('2022-04-01') TO ('2022-05-01');
CREATE TABLE journal_entrys_202205 PARTITION OF journal_entrys FOR VALUES FROM ('2022-05-01') TO ('2022-06-01');
CREATE TABLE journal_entrys_202206 PARTITION OF journal_entrys FOR VALUES FROM ('2022-06-01') TO ('2022-07-01');
CREATE TABLE journal_entrys_202207 PARTITION OF journal_entrys FOR VALUES FROM ('2022-07-01') TO ('2022-08-01');
CREATE TABLE journal_entrys_202208 PARTITION OF journal_entrys FOR VALUES FROM ('2022-08-01') TO ('2022-09-01');
CREATE TABLE journal_entrys_202209 PARTITION OF journal_entrys FOR VALUES FROM ('2022-09-01') TO ('2022-10-01');
CREATE TABLE journal_entrys_202210 PARTITION OF journal_entrys FOR VALUES FROM ('2022-10-01') TO ('2022-11-01');
CREATE TABLE journal_entrys_202211 PARTITION OF journal_entrys FOR VALUES FROM ('2022-11-01') TO ('2022-12-01');
CREATE TABLE journal_entrys_202212 PARTITION OF journal_entrys FOR VALUES FROM ('2022-12-01') TO ('2023-01-01');
CREATE TABLE journal_entrys_202301 PARTITION OF journal_entrys FOR VALUES FROM ('2023-01-01') TO ('2023-02-01');
CREATE TABLE journal_entrys_202302 PARTITION OF journal_entrys FOR VALUES FROM ('2023-02-01') TO ('2023-03-01');
CREATE TABLE journal_entrys_202303 PARTITION OF journal_entrys FOR VALUES FROM ('2023-03-01') TO ('2023-04-01');
CREATE TABLE journal_entrys_202304 PARTITION OF journal_entrys FOR VALUES FROM ('2023-04-01') TO ('2023-05-01');
CREATE TABLE journal_entrys_202305 PARTITION OF journal_entrys FOR VALUES FROM ('2023-05-01') TO ('2023-06-01');
CREATE TABLE journal_entrys_202306 PARTITION OF journal_entrys FOR VALUES FROM ('2023-06-01') TO ('2023-07-01');
CREATE TABLE journal_entrys_202307 PARTITION OF journal_entrys FOR VALUES FROM ('2023-07-01') TO ('2023-08-01');
CREATE TABLE journal_entrys_202308 PARTITION OF journal_entrys FOR VALUES FROM ('2023-08-01') TO ('2023-09-01');
CREATE TABLE journal_entrys_202309 PARTITION OF journal_entrys FOR VALUES FROM ('2023-09-01') TO ('2023-10-01');
CREATE TABLE journal_entrys_202310 PARTITION OF journal_entrys FOR VALUES FROM ('2023-10-01') TO ('2023-11-01');
CREATE TABLE journal_entrys_202311 PARTITION OF journal_entrys FOR VALUES FROM ('2023-11-01') TO ('2023-12-01');
CREATE TABLE journal_entrys_202312 PARTITION OF journal_entrys FOR VALUES FROM ('2023-12-01') TO ('2024-01-01');
CREATE TABLE journal_entrys_202401 PARTITION OF journal_entrys FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE journal_entrys_202402 PARTITION OF journal_entrys FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE journal_entrys_202403 PARTITION OF journal_entrys FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
CREATE TABLE journal_entrys_202404 PARTITION OF journal_entrys FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
CREATE TABLE journal_entrys_202405 PARTITION OF journal_entrys FOR VALUES FROM ('2024-05-01') TO ('2024-06-01');
CREATE TABLE journal_entrys_202406 PARTITION OF journal_entrys FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');
CREATE TABLE journal_entrys_202407 PARTITION OF journal_entrys FOR VALUES FROM ('2024-07-01') TO ('2024-08-01');
CREATE TABLE journal_entrys_202408 PARTITION OF journal_entrys FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
CREATE TABLE journal_entrys_202409 PARTITION OF journal_entrys FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');
CREATE TABLE journal_entrys_202410 PARTITION OF journal_entrys FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE journal_entrys_202411 PARTITION OF journal_entrys FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE journal_entrys_202412 PARTITION OF journal_entrys FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');
CREATE TABLE journal_entrys_202501 PARTITION OF journal_entrys FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE journal_entrys_202502 PARTITION OF journal_entrys FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE journal_entrys_202503 PARTITION OF journal_entrys FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE journal_entrys_202504 PARTITION OF journal_entrys FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE journal_entrys_202505 PARTITION OF journal_entrys FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE journal_entrys_202506 PARTITION OF journal_entrys FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE journal_entrys_202507 PARTITION OF journal_entrys FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE journal_entrys_202508 PARTITION OF journal_entrys FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE journal_entrys_202509 PARTITION OF journal_entrys FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE journal_entrys_202510 PARTITION OF journal_entrys FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE journal_entrys_202511 PARTITION OF journal_entrys FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE journal_entrys_202512 PARTITION OF journal_entrys FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE journal_entrys_202601 PARTITION OF journal_entrys FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE journal_entrys_202602 PARTITION OF journal_entrys FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE journal_entrys_202603 PARTITION OF journal_entrys FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE journal_entrys_202604 PARTITION OF journal_entrys FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE journal_entrys_202605 PARTITION OF journal_entrys FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE journal_entrys_202606 PARTITION OF journal_entrys FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE journal_entrys_202607 PARTITION OF journal_entrys FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE journal_entrys_202608 PARTITION OF journal_entrys FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE journal_entrys_202609 PARTITION OF journal_entrys FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE journal_entrys_202610 PARTITION OF journal_entrys FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE journal_entrys_202611 PARTITION OF journal_entrys FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE journal_entrys_202612 PARTITION OF journal_entrys FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE journal_entrys_202701 PARTITION OF journal_entrys FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE journal_entrys_202702 PARTITION OF journal_entrys FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE journal_entrys_202703 PARTITION OF journal_entrys FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE journal_entrys_202704 PARTITION OF journal_entrys FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE journal_entrys_202705 PARTITION OF journal_entrys FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE journal_entrys_202706 PARTITION OF journal_entrys FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE journal_entrys_202707 PARTITION OF journal_entrys FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE journal_entrys_202708 PARTITION OF journal_entrys FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE journal_entrys_202709 PARTITION OF journal_entrys FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE journal_entrys_202710 PARTITION OF journal_entrys FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE journal_entrys_202711 PARTITION OF journal_entrys FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE journal_entrys_202712 PARTITION OF journal_entrys FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');
CREATE TABLE journal_entrys_202801 PARTITION OF journal_entrys FOR VALUES FROM ('2028-01-01') TO ('2028-02-01');
CREATE TABLE journal_entrys_202802 PARTITION OF journal_entrys FOR VALUES FROM ('2028-02-01') TO ('2028-03-01');
CREATE TABLE journal_entrys_202803 PARTITION OF journal_entrys FOR VALUES FROM ('2028-03-01') TO ('2028-04-01');
CREATE TABLE journal_entrys_202804 PARTITION OF journal_entrys FOR VALUES FROM ('2028-04-01') TO ('2028-05-01');
CREATE TABLE journal_entrys_202805 PARTITION OF journal_entrys FOR VALUES FROM ('2028-05-01') TO ('2028-06-01');
CREATE TABLE journal_entrys_202806 PARTITION OF journal_entrys FOR VALUES FROM ('2028-06-01') TO ('2028-07-01');
CREATE TABLE journal_entrys_202807 PARTITION OF journal_entrys FOR VALUES FROM ('2028-07-01') TO ('2028-08-01');
CREATE TABLE journal_entrys_202808 PARTITION OF journal_entrys FOR VALUES FROM ('2028-08-01') TO ('2028-09-01');
CREATE TABLE journal_entrys_202809 PARTITION OF journal_entrys FOR VALUES FROM ('2028-09-01') TO ('2028-10-01');
CREATE TABLE journal_entrys_202810 PARTITION OF journal_entrys FOR VALUES FROM ('2028-10-01') TO ('2028-11-01');
CREATE TABLE journal_entrys_202811 PARTITION OF journal_entrys FOR VALUES FROM ('2028-11-01') TO ('2028-12-01');
CREATE TABLE journal_entrys_202812 PARTITION OF journal_entrys FOR VALUES FROM ('2028-12-01') TO ('2029-01-01');
CREATE TABLE journal_entrys_202901 PARTITION OF journal_entrys FOR VALUES FROM ('2029-01-01') TO ('2029-02-01');
CREATE TABLE journal_entrys_202902 PARTITION OF journal_entrys FOR VALUES FROM ('2029-02-01') TO ('2029-03-01');
CREATE TABLE journal_entrys_202903 PARTITION OF journal_entrys FOR VALUES FROM ('2029-03-01') TO ('2029-04-01');
CREATE TABLE journal_entrys_202904 PARTITION OF journal_entrys FOR VALUES FROM ('2029-04-01') TO ('2029-05-01');
CREATE TABLE journal_entrys_202905 PARTITION OF journal_entrys FOR VALUES FROM ('2029-05-01') TO ('2029-06-01');
CREATE TABLE journal_entrys_202906 PARTITION OF journal_entrys FOR VALUES FROM ('2029-06-01') TO ('2029-07-01');
CREATE TABLE journal_entrys_202907 PARTITION OF journal_entrys FOR VALUES FROM ('2029-07-01') TO ('2029-08-01');
CREATE TABLE journal_entrys_202908 PARTITION OF journal_entrys FOR VALUES FROM ('2029-08-01') TO ('2029-09-01');
CREATE TABLE journal_entrys_202909 PARTITION OF journal_entrys FOR VALUES FROM ('2029-09-01') TO ('2029-10-01');
CREATE TABLE journal_entrys_202910 PARTITION OF journal_entrys FOR VALUES FROM ('2029-10-01') TO ('2029-11-01');
CREATE TABLE journal_entrys_202911 PARTITION OF journal_entrys FOR VALUES FROM ('2029-11-01') TO ('2029-12-01');
CREATE TABLE journal_entrys_202912 PARTITION OF journal_entrys FOR VALUES FROM ('2029-12-01') TO ('2030-01-01');
CREATE TABLE journal_entrys_203001 PARTITION OF journal_entrys FOR VALUES FROM ('2030-01-01') TO ('2030-02-01');
CREATE TABLE journal_entrys_203002 PARTITION OF journal_entrys FOR VALUES FROM ('2030-02-01') TO ('2030-03-01');
CREATE TABLE journal_entrys_203003 PARTITION OF journal_entrys FOR VALUES FROM ('2030-03-01') TO ('2030-04-01');
CREATE TABLE journal_entrys_203004 PARTITION OF journal_entrys FOR VALUES FROM ('2030-04-01') TO ('2030-05-01');
CREATE TABLE journal_entrys_203005 PARTITION OF journal_entrys FOR VALUES FROM ('2030-05-01') TO ('2030-06-01');
CREATE TABLE journal_entrys_203006 PARTITION OF journal_entrys FOR VALUES FROM ('2030-06-01') TO ('2030-07-01');
CREATE TABLE journal_entrys_203007 PARTITION OF journal_entrys FOR VALUES FROM ('2030-07-01') TO ('2030-08-01');
CREATE TABLE journal_entrys_203008 PARTITION OF journal_entrys FOR VALUES FROM ('2030-08-01') TO ('2030-09-01');
CREATE TABLE journal_entrys_203009 PARTITION OF journal_entrys FOR VALUES FROM ('2030-09-01') TO ('2030-10-01');
CREATE TABLE journal_entrys_203010 PARTITION OF journal_entrys FOR VALUES FROM ('2030-10-01') TO ('2030-11-01');
CREATE TABLE journal_entrys_203011 PARTITION OF journal_entrys FOR VALUES FROM ('2030-11-01') TO ('2030-12-01');
CREATE TABLE journal_entrys_203012 PARTITION OF journal_entrys FOR VALUES FROM ('2030-12-01') TO ('2031-01-01');
CREATE INDEX idx_am_date ON journal_entrys(organization_id, date);
CREATE INDEX idx_am_id ON journal_entrys(id);
CREATE INDEX idx_am_origin ON journal_entrys(origin_type, origin_id);
CREATE INDEX idx_am_reversed ON journal_entrys(reversed_entry_id) WHERE reversed_entry_id IS NOT NULL;
CREATE UNIQUE INDEX idx_am_org_journal_name ON journal_entrys(date, organization_id, journal_id, name) WHERE state != 'cancelled' AND name IS NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_fk_exists() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  fk_val bigint;
  found  boolean;
BEGIN
  fk_val := (to_jsonb(NEW) ->> TG_ARGV[1])::bigint;
  IF fk_val IS NULL THEN
    RETURN NEW;
  END IF;
  EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE id = $1)', TG_ARGV[0])
    INTO found USING fk_val;
  IF NOT found THEN
    RAISE EXCEPTION 'referenced row in % with id % does not exist', TG_ARGV[0], fk_val;
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_aml_move() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  mdate date;
BEGIN
  SELECT date INTO mdate FROM journal_entrys WHERE id = NEW.movement_id;
  IF mdate IS NULL THEN
    RAISE EXCEPTION 'account move % does not exist', NEW.movement_id;
  END IF;
  IF mdate <> NEW.date THEN
    RAISE EXCEPTION 'account move line date % does not match move date %', NEW.date, mdate;
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_journal_entrys_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS sw_partition_moving(tbl name NOT NULL, id bigint NOT NULL, PRIMARY KEY (tbl, id)) ON COMMIT DELETE ROWS;
  IF EXISTS (SELECT 1 FROM sw_partition_moving WHERE tbl = TG_TABLE_NAME AND id = OLD.id) THEN
    RETURN OLD;
  END IF;
  DELETE FROM journal_lines WHERE movement_id = OLD.id;
  IF EXISTS (SELECT 1 FROM journal_entrys WHERE reversed_entry_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by another move', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM asset_depreciation_lines WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by asset_depreciation_lines', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM deferred_schedule_lines WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by deferred_schedule_lines', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM expense_reports WHERE movement_id = OLD.id OR reimbursement_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by expense_reports', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM fixed_assets WHERE original_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by fixed_assets', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM gift_card_transactions WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by gift_card_transactions', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM invoices WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by invoices', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM inbound_costs WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by inbound_costs', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM payments WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by payments', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM payslips WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by payslips', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM cost_layers WHERE journal_entry_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete account move %: referenced by cost_layers', OLD.id;
  END IF;
  RETURN OLD;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION register_partition_move() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS sw_partition_moving(tbl name NOT NULL, id bigint NOT NULL, PRIMARY KEY (tbl, id)) ON COMMIT DELETE ROWS;
  INSERT INTO sw_partition_moving(tbl, id) VALUES (TG_TABLE_NAME, OLD.id) ON CONFLICT DO NOTHING;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION clear_partition_move() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS sw_partition_moving(tbl name NOT NULL, id bigint NOT NULL, PRIMARY KEY (tbl, id)) ON COMMIT DELETE ROWS;
  DELETE FROM sw_partition_moving WHERE tbl = TG_TABLE_NAME AND id = NEW.id;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER enforce_am_reversed BEFORE INSERT OR UPDATE OF reversed_entry_id ON journal_entrys
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'reversed_entry_id');
CREATE TRIGGER enforce_am_delete BEFORE DELETE ON journal_entrys
  FOR EACH ROW EXECUTE FUNCTION enforce_journal_entrys_delete();
CREATE TRIGGER register_am_move BEFORE UPDATE OF date ON journal_entrys
  FOR EACH ROW EXECUTE FUNCTION register_partition_move();
CREATE TRIGGER clear_am_move AFTER INSERT ON journal_entrys
  FOR EACH ROW EXECUTE FUNCTION clear_partition_move();

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_am_reversed ON journal_entrys;
DROP TRIGGER IF EXISTS enforce_am_delete ON journal_entrys;
DROP TRIGGER IF EXISTS register_am_move ON journal_entrys;
DROP TRIGGER IF EXISTS clear_am_move ON journal_entrys;
DROP FUNCTION IF EXISTS enforce_fk_exists();
DROP FUNCTION IF EXISTS enforce_aml_move();
DROP FUNCTION IF EXISTS enforce_journal_entrys_delete();
DROP FUNCTION IF EXISTS register_partition_move();
DROP FUNCTION IF EXISTS clear_partition_move();
DROP INDEX IF EXISTS idx_am_date;
DROP INDEX IF EXISTS idx_am_id;
DROP TABLE journal_entrys;
