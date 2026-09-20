-- +goose Up
SELECT 'up SQL query';
CREATE TABLE stock_movements (
  id BIGINT GENERATED ALWAYS AS IDENTITY,
  organization_id BIGINT,
  shipment_id BIGINT,
  item_id BIGINT NOT NULL,
  qty NUMERIC(18,4) NOT NULL,
  unit_id BIGINT,
  src_location_id BIGINT NOT NULL,
  dst_location_id BIGINT NOT NULL,
  batch_id BIGINT,
  state TEXT CHECK (state IN ('draft','confirmed','assigned','done','cancelled')),
  unit_cost NUMERIC(18,4),
  origin_type TEXT,
  origin_id BIGINT,
  scheduled_date TIMESTAMP,
  date_done TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_stock_movements_organization_id
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_shipment_id
    FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_item_id
    FOREIGN KEY (item_id) REFERENCES item_variants(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_unit_id
    FOREIGN KEY (unit_id) REFERENCES units(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_src_location_id
    FOREIGN KEY (src_location_id) REFERENCES stock_locations(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_dst_location_id
    FOREIGN KEY (dst_location_id) REFERENCES stock_locations(id) ON DELETE RESTRICT,
  CONSTRAINT fk_stock_movements_batch_id
    FOREIGN KEY (batch_id) REFERENCES batches(id) ON DELETE RESTRICT
) PARTITION BY RANGE (date_done);

CREATE TABLE stock_movements_default PARTITION OF stock_movements DEFAULT;
CREATE TABLE stock_movements_202001 PARTITION OF stock_movements FOR VALUES FROM ('2020-01-01') TO ('2020-02-01');
CREATE TABLE stock_movements_202002 PARTITION OF stock_movements FOR VALUES FROM ('2020-02-01') TO ('2020-03-01');
CREATE TABLE stock_movements_202003 PARTITION OF stock_movements FOR VALUES FROM ('2020-03-01') TO ('2020-04-01');
CREATE TABLE stock_movements_202004 PARTITION OF stock_movements FOR VALUES FROM ('2020-04-01') TO ('2020-05-01');
CREATE TABLE stock_movements_202005 PARTITION OF stock_movements FOR VALUES FROM ('2020-05-01') TO ('2020-06-01');
CREATE TABLE stock_movements_202006 PARTITION OF stock_movements FOR VALUES FROM ('2020-06-01') TO ('2020-07-01');
CREATE TABLE stock_movements_202007 PARTITION OF stock_movements FOR VALUES FROM ('2020-07-01') TO ('2020-08-01');
CREATE TABLE stock_movements_202008 PARTITION OF stock_movements FOR VALUES FROM ('2020-08-01') TO ('2020-09-01');
CREATE TABLE stock_movements_202009 PARTITION OF stock_movements FOR VALUES FROM ('2020-09-01') TO ('2020-10-01');
CREATE TABLE stock_movements_202010 PARTITION OF stock_movements FOR VALUES FROM ('2020-10-01') TO ('2020-11-01');
CREATE TABLE stock_movements_202011 PARTITION OF stock_movements FOR VALUES FROM ('2020-11-01') TO ('2020-12-01');
CREATE TABLE stock_movements_202012 PARTITION OF stock_movements FOR VALUES FROM ('2020-12-01') TO ('2021-01-01');
CREATE TABLE stock_movements_202101 PARTITION OF stock_movements FOR VALUES FROM ('2021-01-01') TO ('2021-02-01');
CREATE TABLE stock_movements_202102 PARTITION OF stock_movements FOR VALUES FROM ('2021-02-01') TO ('2021-03-01');
CREATE TABLE stock_movements_202103 PARTITION OF stock_movements FOR VALUES FROM ('2021-03-01') TO ('2021-04-01');
CREATE TABLE stock_movements_202104 PARTITION OF stock_movements FOR VALUES FROM ('2021-04-01') TO ('2021-05-01');
CREATE TABLE stock_movements_202105 PARTITION OF stock_movements FOR VALUES FROM ('2021-05-01') TO ('2021-06-01');
CREATE TABLE stock_movements_202106 PARTITION OF stock_movements FOR VALUES FROM ('2021-06-01') TO ('2021-07-01');
CREATE TABLE stock_movements_202107 PARTITION OF stock_movements FOR VALUES FROM ('2021-07-01') TO ('2021-08-01');
CREATE TABLE stock_movements_202108 PARTITION OF stock_movements FOR VALUES FROM ('2021-08-01') TO ('2021-09-01');
CREATE TABLE stock_movements_202109 PARTITION OF stock_movements FOR VALUES FROM ('2021-09-01') TO ('2021-10-01');
CREATE TABLE stock_movements_202110 PARTITION OF stock_movements FOR VALUES FROM ('2021-10-01') TO ('2021-11-01');
CREATE TABLE stock_movements_202111 PARTITION OF stock_movements FOR VALUES FROM ('2021-11-01') TO ('2021-12-01');
CREATE TABLE stock_movements_202112 PARTITION OF stock_movements FOR VALUES FROM ('2021-12-01') TO ('2022-01-01');
CREATE TABLE stock_movements_202201 PARTITION OF stock_movements FOR VALUES FROM ('2022-01-01') TO ('2022-02-01');
CREATE TABLE stock_movements_202202 PARTITION OF stock_movements FOR VALUES FROM ('2022-02-01') TO ('2022-03-01');
CREATE TABLE stock_movements_202203 PARTITION OF stock_movements FOR VALUES FROM ('2022-03-01') TO ('2022-04-01');
CREATE TABLE stock_movements_202204 PARTITION OF stock_movements FOR VALUES FROM ('2022-04-01') TO ('2022-05-01');
CREATE TABLE stock_movements_202205 PARTITION OF stock_movements FOR VALUES FROM ('2022-05-01') TO ('2022-06-01');
CREATE TABLE stock_movements_202206 PARTITION OF stock_movements FOR VALUES FROM ('2022-06-01') TO ('2022-07-01');
CREATE TABLE stock_movements_202207 PARTITION OF stock_movements FOR VALUES FROM ('2022-07-01') TO ('2022-08-01');
CREATE TABLE stock_movements_202208 PARTITION OF stock_movements FOR VALUES FROM ('2022-08-01') TO ('2022-09-01');
CREATE TABLE stock_movements_202209 PARTITION OF stock_movements FOR VALUES FROM ('2022-09-01') TO ('2022-10-01');
CREATE TABLE stock_movements_202210 PARTITION OF stock_movements FOR VALUES FROM ('2022-10-01') TO ('2022-11-01');
CREATE TABLE stock_movements_202211 PARTITION OF stock_movements FOR VALUES FROM ('2022-11-01') TO ('2022-12-01');
CREATE TABLE stock_movements_202212 PARTITION OF stock_movements FOR VALUES FROM ('2022-12-01') TO ('2023-01-01');
CREATE TABLE stock_movements_202301 PARTITION OF stock_movements FOR VALUES FROM ('2023-01-01') TO ('2023-02-01');
CREATE TABLE stock_movements_202302 PARTITION OF stock_movements FOR VALUES FROM ('2023-02-01') TO ('2023-03-01');
CREATE TABLE stock_movements_202303 PARTITION OF stock_movements FOR VALUES FROM ('2023-03-01') TO ('2023-04-01');
CREATE TABLE stock_movements_202304 PARTITION OF stock_movements FOR VALUES FROM ('2023-04-01') TO ('2023-05-01');
CREATE TABLE stock_movements_202305 PARTITION OF stock_movements FOR VALUES FROM ('2023-05-01') TO ('2023-06-01');
CREATE TABLE stock_movements_202306 PARTITION OF stock_movements FOR VALUES FROM ('2023-06-01') TO ('2023-07-01');
CREATE TABLE stock_movements_202307 PARTITION OF stock_movements FOR VALUES FROM ('2023-07-01') TO ('2023-08-01');
CREATE TABLE stock_movements_202308 PARTITION OF stock_movements FOR VALUES FROM ('2023-08-01') TO ('2023-09-01');
CREATE TABLE stock_movements_202309 PARTITION OF stock_movements FOR VALUES FROM ('2023-09-01') TO ('2023-10-01');
CREATE TABLE stock_movements_202310 PARTITION OF stock_movements FOR VALUES FROM ('2023-10-01') TO ('2023-11-01');
CREATE TABLE stock_movements_202311 PARTITION OF stock_movements FOR VALUES FROM ('2023-11-01') TO ('2023-12-01');
CREATE TABLE stock_movements_202312 PARTITION OF stock_movements FOR VALUES FROM ('2023-12-01') TO ('2024-01-01');
CREATE TABLE stock_movements_202401 PARTITION OF stock_movements FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE stock_movements_202402 PARTITION OF stock_movements FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE stock_movements_202403 PARTITION OF stock_movements FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
CREATE TABLE stock_movements_202404 PARTITION OF stock_movements FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
CREATE TABLE stock_movements_202405 PARTITION OF stock_movements FOR VALUES FROM ('2024-05-01') TO ('2024-06-01');
CREATE TABLE stock_movements_202406 PARTITION OF stock_movements FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');
CREATE TABLE stock_movements_202407 PARTITION OF stock_movements FOR VALUES FROM ('2024-07-01') TO ('2024-08-01');
CREATE TABLE stock_movements_202408 PARTITION OF stock_movements FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
CREATE TABLE stock_movements_202409 PARTITION OF stock_movements FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');
CREATE TABLE stock_movements_202410 PARTITION OF stock_movements FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE stock_movements_202411 PARTITION OF stock_movements FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE stock_movements_202412 PARTITION OF stock_movements FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');
CREATE TABLE stock_movements_202501 PARTITION OF stock_movements FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE stock_movements_202502 PARTITION OF stock_movements FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE stock_movements_202503 PARTITION OF stock_movements FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE stock_movements_202504 PARTITION OF stock_movements FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE stock_movements_202505 PARTITION OF stock_movements FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE stock_movements_202506 PARTITION OF stock_movements FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE stock_movements_202507 PARTITION OF stock_movements FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE stock_movements_202508 PARTITION OF stock_movements FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE stock_movements_202509 PARTITION OF stock_movements FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE stock_movements_202510 PARTITION OF stock_movements FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE stock_movements_202511 PARTITION OF stock_movements FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE stock_movements_202512 PARTITION OF stock_movements FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE stock_movements_202601 PARTITION OF stock_movements FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE stock_movements_202602 PARTITION OF stock_movements FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE stock_movements_202603 PARTITION OF stock_movements FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE stock_movements_202604 PARTITION OF stock_movements FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE stock_movements_202605 PARTITION OF stock_movements FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE stock_movements_202606 PARTITION OF stock_movements FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE stock_movements_202607 PARTITION OF stock_movements FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE stock_movements_202608 PARTITION OF stock_movements FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE stock_movements_202609 PARTITION OF stock_movements FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE stock_movements_202610 PARTITION OF stock_movements FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE stock_movements_202611 PARTITION OF stock_movements FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE stock_movements_202612 PARTITION OF stock_movements FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE stock_movements_202701 PARTITION OF stock_movements FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE stock_movements_202702 PARTITION OF stock_movements FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE stock_movements_202703 PARTITION OF stock_movements FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE stock_movements_202704 PARTITION OF stock_movements FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE stock_movements_202705 PARTITION OF stock_movements FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE stock_movements_202706 PARTITION OF stock_movements FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE stock_movements_202707 PARTITION OF stock_movements FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE stock_movements_202708 PARTITION OF stock_movements FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE stock_movements_202709 PARTITION OF stock_movements FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE stock_movements_202710 PARTITION OF stock_movements FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE stock_movements_202711 PARTITION OF stock_movements FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE stock_movements_202712 PARTITION OF stock_movements FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');
CREATE TABLE stock_movements_202801 PARTITION OF stock_movements FOR VALUES FROM ('2028-01-01') TO ('2028-02-01');
CREATE TABLE stock_movements_202802 PARTITION OF stock_movements FOR VALUES FROM ('2028-02-01') TO ('2028-03-01');
CREATE TABLE stock_movements_202803 PARTITION OF stock_movements FOR VALUES FROM ('2028-03-01') TO ('2028-04-01');
CREATE TABLE stock_movements_202804 PARTITION OF stock_movements FOR VALUES FROM ('2028-04-01') TO ('2028-05-01');
CREATE TABLE stock_movements_202805 PARTITION OF stock_movements FOR VALUES FROM ('2028-05-01') TO ('2028-06-01');
CREATE TABLE stock_movements_202806 PARTITION OF stock_movements FOR VALUES FROM ('2028-06-01') TO ('2028-07-01');
CREATE TABLE stock_movements_202807 PARTITION OF stock_movements FOR VALUES FROM ('2028-07-01') TO ('2028-08-01');
CREATE TABLE stock_movements_202808 PARTITION OF stock_movements FOR VALUES FROM ('2028-08-01') TO ('2028-09-01');
CREATE TABLE stock_movements_202809 PARTITION OF stock_movements FOR VALUES FROM ('2028-09-01') TO ('2028-10-01');
CREATE TABLE stock_movements_202810 PARTITION OF stock_movements FOR VALUES FROM ('2028-10-01') TO ('2028-11-01');
CREATE TABLE stock_movements_202811 PARTITION OF stock_movements FOR VALUES FROM ('2028-11-01') TO ('2028-12-01');
CREATE TABLE stock_movements_202812 PARTITION OF stock_movements FOR VALUES FROM ('2028-12-01') TO ('2029-01-01');
CREATE TABLE stock_movements_202901 PARTITION OF stock_movements FOR VALUES FROM ('2029-01-01') TO ('2029-02-01');
CREATE TABLE stock_movements_202902 PARTITION OF stock_movements FOR VALUES FROM ('2029-02-01') TO ('2029-03-01');
CREATE TABLE stock_movements_202903 PARTITION OF stock_movements FOR VALUES FROM ('2029-03-01') TO ('2029-04-01');
CREATE TABLE stock_movements_202904 PARTITION OF stock_movements FOR VALUES FROM ('2029-04-01') TO ('2029-05-01');
CREATE TABLE stock_movements_202905 PARTITION OF stock_movements FOR VALUES FROM ('2029-05-01') TO ('2029-06-01');
CREATE TABLE stock_movements_202906 PARTITION OF stock_movements FOR VALUES FROM ('2029-06-01') TO ('2029-07-01');
CREATE TABLE stock_movements_202907 PARTITION OF stock_movements FOR VALUES FROM ('2029-07-01') TO ('2029-08-01');
CREATE TABLE stock_movements_202908 PARTITION OF stock_movements FOR VALUES FROM ('2029-08-01') TO ('2029-09-01');
CREATE TABLE stock_movements_202909 PARTITION OF stock_movements FOR VALUES FROM ('2029-09-01') TO ('2029-10-01');
CREATE TABLE stock_movements_202910 PARTITION OF stock_movements FOR VALUES FROM ('2029-10-01') TO ('2029-11-01');
CREATE TABLE stock_movements_202911 PARTITION OF stock_movements FOR VALUES FROM ('2029-11-01') TO ('2029-12-01');
CREATE TABLE stock_movements_202912 PARTITION OF stock_movements FOR VALUES FROM ('2029-12-01') TO ('2030-01-01');
CREATE TABLE stock_movements_203001 PARTITION OF stock_movements FOR VALUES FROM ('2030-01-01') TO ('2030-02-01');
CREATE TABLE stock_movements_203002 PARTITION OF stock_movements FOR VALUES FROM ('2030-02-01') TO ('2030-03-01');
CREATE TABLE stock_movements_203003 PARTITION OF stock_movements FOR VALUES FROM ('2030-03-01') TO ('2030-04-01');
CREATE TABLE stock_movements_203004 PARTITION OF stock_movements FOR VALUES FROM ('2030-04-01') TO ('2030-05-01');
CREATE TABLE stock_movements_203005 PARTITION OF stock_movements FOR VALUES FROM ('2030-05-01') TO ('2030-06-01');
CREATE TABLE stock_movements_203006 PARTITION OF stock_movements FOR VALUES FROM ('2030-06-01') TO ('2030-07-01');
CREATE TABLE stock_movements_203007 PARTITION OF stock_movements FOR VALUES FROM ('2030-07-01') TO ('2030-08-01');
CREATE TABLE stock_movements_203008 PARTITION OF stock_movements FOR VALUES FROM ('2030-08-01') TO ('2030-09-01');
CREATE TABLE stock_movements_203009 PARTITION OF stock_movements FOR VALUES FROM ('2030-09-01') TO ('2030-10-01');
CREATE TABLE stock_movements_203010 PARTITION OF stock_movements FOR VALUES FROM ('2030-10-01') TO ('2030-11-01');
CREATE TABLE stock_movements_203011 PARTITION OF stock_movements FOR VALUES FROM ('2030-11-01') TO ('2030-12-01');
CREATE TABLE stock_movements_203012 PARTITION OF stock_movements FOR VALUES FROM ('2030-12-01') TO ('2031-01-01');
CREATE INDEX idx_move_product_date ON stock_movements(item_id, date_done);
CREATE INDEX idx_move_locations ON stock_movements(src_location_id, dst_location_id);
CREATE INDEX idx_move_origin ON stock_movements(origin_type, origin_id);
CREATE INDEX idx_sm_id ON stock_movements(id);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_stock_movements_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS sw_partition_moving(tbl name NOT NULL, id bigint NOT NULL, PRIMARY KEY (tbl, id)) ON COMMIT DELETE ROWS;
  IF EXISTS (SELECT 1 FROM sw_partition_moving WHERE tbl = TG_TABLE_NAME AND id = OLD.id) THEN
    RETURN OLD;
  END IF;
  IF EXISTS (SELECT 1 FROM stock_holds WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by stock_holds', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM cost_layers WHERE movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by cost_layers', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM rma_lines WHERE stock_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by rma_lines', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM mo_components WHERE stock_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by mo_components', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM inbound_cost_adjustments WHERE stock_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by inbound_cost_adjustments', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM service_order_lines WHERE stock_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by service_order_lines', OLD.id;
  END IF;
  IF EXISTS (SELECT 1 FROM dropship_links WHERE stock_movement_id = OLD.id) THEN
    RAISE EXCEPTION 'cannot delete stock move %: referenced by dropship_links', OLD.id;
  END IF;
  RETURN OLD;
END $$;
-- +goose StatementEnd

CREATE TRIGGER enforce_sm_delete BEFORE DELETE ON stock_movements
  FOR EACH ROW EXECUTE FUNCTION enforce_stock_movements_delete();
CREATE TRIGGER register_sm_move BEFORE UPDATE OF date_done ON stock_movements
  FOR EACH ROW EXECUTE FUNCTION register_partition_move();
CREATE TRIGGER clear_sm_move AFTER INSERT ON stock_movements
  FOR EACH ROW EXECUTE FUNCTION clear_partition_move();

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_sm_delete ON stock_movements;
DROP TRIGGER IF EXISTS register_sm_move ON stock_movements;
DROP TRIGGER IF EXISTS clear_sm_move ON stock_movements;
DROP FUNCTION IF EXISTS enforce_stock_movements_delete();
DROP INDEX IF EXISTS idx_move_product_date;
DROP INDEX IF EXISTS idx_move_locations;
DROP INDEX IF EXISTS idx_move_origin;
DROP INDEX IF EXISTS idx_sm_id;
DROP TABLE stock_movements;
