-- +goose Up
SELECT 'up SQL query';
CREATE TABLE deferred_schedules (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  type TEXT CHECK (type IN ('deferred_revenue','deferred_expense','prepaid')),
  source_type TEXT,
  source_id BIGINT,
  contact_id BIGINT,
  item_id BIGINT,
  total_amount NUMERIC(18,4),
  recognized_amount NUMERIC(18,4) DEFAULT 0,
  balance_sheet_account_id BIGINT,
  pl_account_id BIGINT,
  method TEXT CHECK (method IN ('linear','manual','milestone')),
  date_start DATE,
  date_end DATE,
  periods INT,
  state TEXT CHECK (state IN ('draft','running','done','cancelled')),
  dimension_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_deferred_schedules_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_deferred_schedules_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_deferred_schedules_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_deferred_schedules_balance_sheet_account_id
    FOREIGN KEY (balance_sheet_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_deferred_schedules_pl_account_id
    FOREIGN KEY (pl_account_id)
    REFERENCES accounts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_deferred_schedules_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE deferred_schedules;
