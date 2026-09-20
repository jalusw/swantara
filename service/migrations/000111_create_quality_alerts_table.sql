-- +goose Up
SELECT 'up SQL query';
CREATE TABLE quality_alerts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id BIGINT,
  batch_id BIGINT,
  check_id BIGINT,
  title TEXT,
  description TEXT,
  severity TEXT,
  state TEXT CHECK (state IN ('open','in_progress','solved','cancelled')),
  assigned_to BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_quality_alerts_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_alerts_batch_id
    FOREIGN KEY (batch_id)
    REFERENCES batches(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_alerts_check_id
    FOREIGN KEY (check_id)
    REFERENCES quality_checks(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_quality_alerts_assigned_to
    FOREIGN KEY (assigned_to)
    REFERENCES users(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE quality_alerts;
