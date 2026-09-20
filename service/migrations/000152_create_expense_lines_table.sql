-- +goose Up
SELECT 'up SQL query';
CREATE TABLE expense_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  report_id BIGINT,
  employee_id BIGINT,
  category_id BIGINT,
  item_id BIGINT,
  description TEXT,
  expense_date DATE,
  quantity NUMERIC(18,4) DEFAULT 1,
  unit_price NUMERIC(18,4),
  amount NUMERIC(18,4),
  tax_ids BIGINT[],
  currency_code CHAR(3),
  dimension_id BIGINT,
  project_id BIGINT,
  reimbursable BOOLEAN DEFAULT true,
  receipt_attachment_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_expense_lines_report_id
    FOREIGN KEY (report_id)
    REFERENCES expense_reports(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_category_id
    FOREIGN KEY (category_id)
    REFERENCES expense_categories(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_dimension_id
    FOREIGN KEY (dimension_id)
    REFERENCES dimensions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_project_id
    FOREIGN KEY (project_id)
    REFERENCES projects(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_expense_lines_receipt_attachment_id
    FOREIGN KEY (receipt_attachment_id)
    REFERENCES attachments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE expense_lines;
