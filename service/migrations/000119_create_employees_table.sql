-- +goose Up
SELECT 'up SQL query';
CREATE TABLE employees (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  contact_id BIGINT,
  user_id BIGINT,
  employee_number TEXT UNIQUE,
  department_id BIGINT,
  job_position_id BIGINT,
  manager_id BIGINT,
  hire_date DATE,
  termination_date DATE,
  employment_type TEXT,
  work_location TEXT,
  active BOOLEAN DEFAULT true,
  requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_employees_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employees_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employees_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employees_department_id
    FOREIGN KEY (department_id)
    REFERENCES departments(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employees_job_position_id
    FOREIGN KEY (job_position_id)
    REFERENCES job_positions(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_employees_manager_id
    FOREIGN KEY (manager_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE employees;
