-- +goose Up
SELECT 'up SQL query';
CREATE TABLE deferred_schedule_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  schedule_id BIGINT,
  sequence INT,
  recognition_date DATE,
  amount NUMERIC(18,4),
  movement_id BIGINT,
  posted BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_deferred_schedule_lines_schedule_id
    FOREIGN KEY (schedule_id)
    REFERENCES deferred_schedules(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

CREATE TRIGGER enforce_dsl_move BEFORE INSERT OR UPDATE OF movement_id ON deferred_schedule_lines
  FOR EACH ROW EXECUTE FUNCTION enforce_fk_exists('journal_entrys', 'movement_id');

-- +goose Down
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS enforce_dsl_move ON deferred_schedule_lines;
DROP TABLE deferred_schedule_lines;
