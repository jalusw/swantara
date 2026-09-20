-- +goose Up
SELECT 'up SQL query';
CREATE TABLE employment_contracts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id BIGINT,
  date_start DATE,
  date_end DATE,
  wage NUMERIC(18,4),
  wage_type TEXT,
  currency_code CHAR(3),
  state TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_employment_contracts_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employment_contracts_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE employment_contracts;
