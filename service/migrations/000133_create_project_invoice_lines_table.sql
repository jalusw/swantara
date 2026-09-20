-- +goose Up
SELECT 'up SQL query';
CREATE TABLE project_invoice_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  project_id BIGINT,
  invoice_id BIGINT,
  invoice_line_id BIGINT,
  timesheet_id BIGINT,
  qty NUMERIC(18,4),
  unit_price NUMERIC(18,4),
  amount NUMERIC(18,4),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_project_invoice_lines_project_id
    FOREIGN KEY (project_id)
    REFERENCES projects(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_project_invoice_lines_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_project_invoice_lines_timesheet_id
    FOREIGN KEY (timesheet_id)
    REFERENCES timesheets(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE project_invoice_lines;
