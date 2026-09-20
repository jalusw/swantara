-- +goose Up
SELECT 'up SQL query';
CREATE TABLE leave_requests (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id BIGINT,
  leave_type_id BIGINT,
  date_from DATE,
  date_to DATE,
  days NUMERIC(6,2),
  state TEXT CHECK (state IN ('draft','submitted','approved','refused')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_leave_requests_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_leave_requests_leave_type_id
    FOREIGN KEY (leave_type_id)
    REFERENCES leave_types(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE leave_requests;
