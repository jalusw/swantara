-- +goose Up
SELECT 'up SQL query';
CREATE TABLE attendances (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id BIGINT NOT NULL,
  organization_id BIGINT NOT NULL,
  check_in TIMESTAMP,
  check_out TIMESTAMP,
  worked_hours NUMERIC(8,2),
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  attendance_type VARCHAR(20) NOT NULL DEFAULT 'normal',
  source VARCHAR(30) NOT NULL DEFAULT 'manual',
  break_minutes INTEGER NOT NULL DEFAULT 0,
  overtime_hours NUMERIC(8,2) NOT NULL DEFAULT 0,
  late_minutes INTEGER NOT NULL DEFAULT 0,
  early_departure_minutes INTEGER NOT NULL DEFAULT 0,
  notes TEXT,
  confirmed_at TIMESTAMP,
  shift_id BIGINT,
  latitude NUMERIC(10,7),
  longitude NUMERIC(10,7),
  device_id VARCHAR(255),
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_attendances_employee_id
    FOREIGN KEY (employee_id)
    REFERENCES employees(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_attendances_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_attendances_shift_id
    FOREIGN KEY (shift_id)
    REFERENCES shifts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_attendances_employee_id ON attendances(employee_id);
CREATE INDEX idx_attendances_organization_id ON attendances(organization_id);
CREATE INDEX idx_attendances_check_in ON attendances(check_in);
CREATE INDEX idx_attendances_employee_check_in ON attendances(employee_id, check_in);
CREATE INDEX idx_attendances_shift_id ON attendances (shift_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_attendances_employee_checkin_date
    ON attendances (employee_id, (DATE(check_in)))
    WHERE deleted_at IS NULL;

-- +goose Down
SELECT 'down SQL query';
DROP TABLE attendances;
